package meta

import (
	"context"
	"net/url"
	"regexp"
	"strings"
)

type gogProduct struct {
	Title  string `json:"title"`
	Images struct {
		Background string `json:"background"`
		Logo2x     string `json:"logo2x"`
		Icon       string `json:"icon"`
	} `json:"images"`
	Description struct {
		Lead string `json:"lead"`
		Full string `json:"full"`
	} `json:"description"`
	ReleaseDate string `json:"release_date"`
}

var reGogID = regexp.MustCompile(`^\d{1,20}$`)

func (c *Client) gogProduct(ctx context.Context, id string) (*gogProduct, error) {
	if !reGogID.MatchString(id) {
		return nil, errNotFound
	}
	var p gogProduct
	if err := c.getJSON(ctx, "https://api.gog.com/products/"+url.PathEscape(id)+"?expand=description", &p, nil); err != nil {
		return nil, err
	}
	return &p, nil
}

// gogURL completes GOG's protocol-relative image links.
func gogURL(s string) string {
	switch {
	case s == "":
		return ""
	case strings.HasPrefix(s, "//"):
		return "https:" + s
	case strings.HasPrefix(s, "https://"):
		return s
	}
	return ""
}
