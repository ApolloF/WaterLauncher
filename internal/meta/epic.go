package meta

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Epic's catalog knows the games only Epic sells (and every other game in
// someone's Epic library): their art, description and developer. It's
// read with the public launcher client's own token (no sign-in), as open
// source launchers do.
const (
	epicClientID     = "34a02cf8f4414e29b15921876da36f9a"
	epicClientSecret = "daafbccc737745039dffe53d94fc76cf"
	epicTokenURL     = "https://account-public-service-prod03.ol.epicgames.com/account/api/oauth/token"
	epicCatalogURL   = "https://catalog-public-service-prod06.ol.epicgames.com/catalog/api/shared/namespace/"
)

type epicItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Developer   string `json:"developer"`
	KeyImages   []struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"keyImages"`
	ReleaseInfo []struct {
		DateAdded string `json:"dateAdded"`
	} `json:"releaseInfo"`
}

// image returns the first key image of these types.
func (e *epicItem) image(types ...string) string {
	for _, t := range types {
		for _, k := range e.KeyImages {
			if k.Type == t && strings.HasPrefix(k.URL, "https://") {
				return k.URL
			}
		}
	}
	return ""
}

var reEpicID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// epicToken returns an app token for the catalog, kept until it expires.
func (c *Client) epicToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.epicTok != "" && time.Now().Before(c.epicExp) {
		t := c.epicTok
		c.mu.Unlock()
		return t, nil
	}
	c.mu.Unlock()
	u, _ := url.Parse(epicTokenURL)
	if !allowed(u) {
		return "", ErrNotAllowed
	}
	form := url.Values{"grant_type": {"client_credentials"}, "token_type": {"eg1"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, epicTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "basic "+base64.StdEncoding.EncodeToString([]byte(epicClientID+":"+epicClientSecret)))
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Epic: %s", resp.Status)
	}
	var t struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&t); err != nil {
		return "", err
	}
	if t.AccessToken == "" {
		return "", errors.New("Epic: no token")
	}
	c.mu.Lock()
	c.epicTok, c.epicExp = t.AccessToken, time.Now().Add(time.Duration(max(60, t.ExpiresIn-60))*time.Second)
	c.mu.Unlock()
	return t.AccessToken, nil
}

// epicGame reads a game from Epic's catalog; app is "namespace:item:appName"
// as the Epic scan records it.
func (c *Client) epicGame(ctx context.Context, app string) (*epicItem, error) {
	parts := strings.Split(app, ":")
	if len(parts) < 2 || !reEpicID.MatchString(parts[0]) || !reEpicID.MatchString(parts[1]) {
		return nil, errNotFound
	}
	tok, err := c.epicToken(ctx)
	if err != nil {
		return nil, err
	}
	var resp map[string]epicItem
	u := epicCatalogURL + url.PathEscape(parts[0]) + "/bulk/items?id=" + url.QueryEscape(parts[1]) + "&country=US&locale=en-US&includeMainGameDetails=true"
	if err := c.getJSON(ctx, u, &resp, map[string]string{"Authorization": "bearer " + tok}); err != nil {
		if errors.Is(err, ErrRateLimited) {
			// Epic answers 403 for items it won't show; that isn't worth waiting for.
			return nil, errNotFound
		}
		return nil, err
	}
	e, ok := resp[parts[1]]
	if !ok || e.Title == "" {
		return nil, errNotFound
	}
	return &e, nil
}
