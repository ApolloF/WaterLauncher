package owned

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGOG(t *testing.T) {
	games, err := GOG("testdata/galaxy-2.0.db")
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]string{}
	for _, g := range games {
		byID[g.ID] = g.Title
		if g.ID == "1453375253" && (g.Playtime != 125*60 || g.LastPlayed == 0 || g.InstallURI != "goggalaxy://openGameView/gog_1453375253") {
			t.Errorf("witcher = %+v", g)
		}
	}
	// Steam releases in Galaxy aren't GOG games; a GOG release without a title is skipped.
	if len(byID) != 2 || byID["1207658924"] != "Quiet Harbor" || byID["1453375253"] != "The Witcher 3: Wild Hunt" {
		t.Errorf("games = %v", byID)
	}
	if _, err := GOG("testdata/missing.db"); err == nil {
		t.Error("a missing database must fail")
	}
}

func TestSteam(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		switch r.URL.Query().Get("steamid") {
		case "private":
			w.Write([]byte(`{"response":{}}`))
		case "badkey":
			w.WriteHeader(403)
		default:
			w.Write([]byte(`{"response":{"game_count":2,"games":[
				{"appid":620,"name":"Portal 2","playtime_forever":90,"rtime_last_played":1700000000},
				{"appid":0,"name":"broken"}]}}`))
		}
	}))
	defer srv.Close()
	c := NewClient()
	c.base = map[string]string{"api.steampowered.com": srv.URL}
	key := strings.Repeat("A1", 16)
	games, err := c.Steam(context.Background(), key, "7656")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 1 || games[0].ID != "620" || games[0].Playtime != 5400 || games[0].InstallURI != "steam://install/620" {
		t.Errorf("games = %+v", games)
	}
	if !strings.Contains(gotQuery, "include_appinfo=1") {
		t.Errorf("query = %s", gotQuery)
	}
	if _, err := c.Steam(context.Background(), key, "private"); err != ErrSteamPrivate {
		t.Errorf("private profile = %v", err)
	}
	if _, err := c.Steam(context.Background(), key, "badkey"); err == nil || !strings.Contains(err.Error(), "key") {
		t.Errorf("bad key = %v", err)
	}
	if _, err := c.Steam(context.Background(), "nope", "1"); err == nil {
		t.Error("a malformed key must fail before any request")
	}
}

func TestAllowlist(t *testing.T) {
	c := NewClient()
	var v any
	if err := c.do(context.Background(), "GET", "https://example.com/x", nil, nil, &v); err == nil {
		t.Error("other hosts must be refused")
	}
	if err := c.do(context.Background(), "GET", "http://api.steampowered.com/x", nil, nil, &v); err == nil {
		t.Error("plain http must be refused")
	}
}

func TestEpicCode(t *testing.T) {
	code := "0123456789abcdef0123456789abcdef"
	for _, in := range []string{code, "  " + code + "\n", `{"warning":"x","redirectUrl":"https://…","authorizationCode":"` + code + `","sid":null}`} {
		if got, ok := EpicCode(in); !ok || got != code {
			t.Errorf("EpicCode(%q) = %q, %v", in, got, ok)
		}
	}
	if _, ok := EpicCode("nope"); ok {
		t.Error("junk accepted")
	}
	if !strings.HasPrefix(EpicLoginURL, "https://www.epicgames.com/id/login?redirectUrl=") {
		t.Error(EpicLoginURL)
	}
}

func TestEpic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/account/api/oauth/token":
			r.ParseForm()
			if user, _, ok := r.BasicAuth(); !ok || user != epicClientID {
				// Epic's header is "basic …" in lower case; BasicAuth accepts it.
				w.WriteHeader(401)
				return
			}
			if r.Form.Get("code") == "bad" {
				w.WriteHeader(400)
				return
			}
			w.Write([]byte(`{"access_token":"AT","refresh_token":"RT","account_id":"acc","displayName":"Me","expires_at":"2030-01-01T00:00:00.000Z"}`))
		case r.URL.Path == "/library/api/public/items":
			if r.Header.Get("Authorization") != "bearer AT" {
				w.WriteHeader(401)
				return
			}
			if r.URL.Query().Get("cursor") == "" {
				w.Write([]byte(`{"records":[
					{"namespace":"ns1","catalogItemId":"game1","appName":"Fortress"},
					{"namespace":"ns1","catalogItemId":"dlc1","appName":"FortressDLC"},
					{"namespace":"ue","catalogItemId":"asset","appName":"Asset"}],
					"responseMetadata":{"nextCursor":"p2"}}`))
			} else {
				w.Write([]byte(`{"records":[{"namespace":"ns2","catalogItemId":"app2","appName":"Launcher"}],"responseMetadata":{}}`))
			}
		case strings.HasPrefix(r.URL.Path, "/catalog/api/shared/namespace/ns1/bulk/items"):
			w.Write([]byte(`{"game1":{"title":"Fortress","categories":[{"path":"games"},{"path":"applications"}]},
				"dlc1":{"title":"Fortress DLC","categories":[{"path":"games"}],"mainGameItem":{"id":"game1"}}}`))
		case strings.HasPrefix(r.URL.Path, "/catalog/api/shared/namespace/ns2/bulk/items"):
			w.Write([]byte(`{"app2":{"title":"Some Tool","categories":[{"path":"applications"}]}}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	c := NewClient()
	c.base = map[string]string{
		"account-public-service-prod03.ol.epicgames.com": srv.URL,
		"library-service.live.use1a.on.epicgames.com":    srv.URL,
		"catalog-public-service-prod06.ol.epicgames.com": srv.URL,
	}
	ctx := context.Background()
	tok, err := c.EpicSignIn(ctx, "good")
	if err != nil || tok.RefreshToken != "RT" || tok.DisplayName != "Me" {
		t.Fatalf("sign in = %+v, %v", tok, err)
	}
	if _, err := c.EpicSignIn(ctx, "bad"); err == nil || !strings.Contains(err.Error(), "sign in again") {
		t.Errorf("bad code = %v", err)
	}
	games, err := c.Epic(ctx, tok.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 1 || games[0].Title != "Fortress" || games[0].ID != "ns1:game1:Fortress" ||
		games[0].InstallURI != "com.epicgames.launcher://apps/ns1:game1:Fortress?action=install" {
		t.Errorf("games = %+v", games)
	}
}

// TestSteamIDLive checks the account lookup on this PC when OWNED_LIVE is set.
func TestSteamIDLive(t *testing.T) {
	if os.Getenv("OWNED_LIVE") == "" {
		t.Skip("OWNED_LIVE not set")
	}
	id, err := SteamID()
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 17 || !strings.HasPrefix(id, "7656119") {
		t.Errorf("steamID64 %q looks wrong", id)
	}
	t.Logf("GOG Galaxy database: %q", GalaxyDB())
}

func TestSteamBadKeyLive(t *testing.T) {
	if os.Getenv("OWNED_LIVE") == "" {
		t.Skip("OWNED_LIVE not set")
	}
	id, _ := SteamID()
	_, err := NewClient().Steam(context.Background(), strings.Repeat("0", 32), id)
	t.Logf("invalid key: %v", err)
	if err == nil || strings.Contains(err.Error(), "0000") {
		t.Errorf("want a key error that doesn't echo the key, got %v", err)
	}
}
