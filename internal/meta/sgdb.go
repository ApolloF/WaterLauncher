package meta

import (
	"context"
	"net/url"
	"strconv"
)

// SteamGridDB needs a (free) personal API key; without one it is skipped.
const sgdbAPI = "https://www.steamgriddb.com/api/v2/"

type sgdbImage struct {
	ID    int    `json:"id"`
	URL   string `json:"url"`
	Thumb string `json:"thumb"`
	Width int    `json:"width"`
	Score int    `json:"score"`
	Nsfw  bool   `json:"nsfw"`
	Humor bool   `json:"humor"`
}

func (c *Client) sgdb(ctx context.Context, path string, v any) error {
	key := ""
	if c.sgdbKey != nil {
		key = c.sgdbKey()
	}
	if key == "" {
		return errNotFound
	}
	return c.getJSON(ctx, sgdbAPI+path, v, map[string]string{"Authorization": "Bearer " + key})
}

// sgdbGameID finds SteamGridDB's id for a game, by Steam app id or by title.
func (c *Client) sgdbGameID(ctx context.Context, steamAppID int, title string) (int, error) {
	var resp struct {
		Success bool  `json:"success"`
		Data    json1 `json:"data"`
	}
	if steamAppID > 0 {
		if err := c.sgdb(ctx, "games/steam/"+strconv.Itoa(steamAppID), &resp); err == nil && resp.Success && resp.Data.ID > 0 {
			return resp.Data.ID, nil
		}
	}
	if title == "" {
		return 0, errNotFound
	}
	var search struct {
		Success bool    `json:"success"`
		Data    []json1 `json:"data"`
	}
	if err := c.sgdb(ctx, "search/autocomplete/"+url.PathEscape(title), &search); err != nil {
		return 0, err
	}
	if !search.Success || len(search.Data) == 0 {
		return 0, errNotFound
	}
	return search.Data[0].ID, nil
}

type json1 struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// sgdbArt returns SteamGridDB images of one kind: "grids" (covers),
// "heroes", "logos" or "icons", best-rated first, without NSFW or joke art.
func (c *Client) sgdbArt(ctx context.Context, gameID int, kind string) ([]sgdbImage, error) {
	q := ""
	switch kind {
	case "grids":
		q = "?dimensions=600x900&types=static"
	case "heroes", "logos", "icons":
		q = "?types=static"
	default:
		return nil, errNotFound
	}
	var resp struct {
		Success bool        `json:"success"`
		Data    []sgdbImage `json:"data"`
	}
	if err := c.sgdb(ctx, kind+"/game/"+strconv.Itoa(gameID)+q, &resp); err != nil {
		return nil, err
	}
	var out []sgdbImage
	for _, im := range resp.Data {
		if !im.Nsfw && !im.Humor && im.URL != "" {
			out = append(out, im)
		}
	}
	return out, nil
}
