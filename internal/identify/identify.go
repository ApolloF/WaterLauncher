package identify

import (
	"path/filepath"
	"strconv"

	"github.com/ApolloF/WaterLauncher/internal/scan"
)

// Index answers lookups by id, title and install folder name.
type Index struct {
	bySteam map[int]*Entry
	byGog   map[string]*Entry
	byName  map[string]*Entry   // normalized title or alias
	byLoose map[string]*Entry   // scan.LooseKey of titles and aliases; nil when two games share one
	byDir   map[string][]*Entry // normalized install folder name
}

func build(es []Entry) *Index {
	ix := &Index{
		bySteam: make(map[int]*Entry, len(es)/2),
		byGog:   make(map[string]*Entry, len(es)/8),
		byName:  make(map[string]*Entry, len(es)),
		byLoose: make(map[string]*Entry, len(es)),
		byDir:   make(map[string][]*Entry, len(es)/2),
	}
	for i := range es {
		e := &es[i]
		if e.SteamID > 0 {
			if _, ok := ix.bySteam[e.SteamID]; !ok {
				ix.bySteam[e.SteamID] = e
			}
		}
		if e.GogID != "" {
			ix.byGog[e.GogID] = e
		}
		for _, n := range append([]string{e.Name}, e.Aliases...) {
			if k := scan.Normalize(n); len(k) >= 2 {
				if _, ok := ix.byName[k]; !ok {
					ix.byName[k] = e
				}
			}
			if k := scan.LooseKey(n); len(k) >= 4 {
				if prev, ok := ix.byLoose[k]; !ok {
					ix.byLoose[k] = e
				} else if prev != nil && prev != e && (prev.SteamID == 0 || prev.SteamID != e.SteamID) {
					ix.byLoose[k] = nil // ambiguous: better no match than a wrong one
				}
			}
		}
		for _, d := range e.InstallDirs {
			if k := scan.Normalize(d); len(k) >= 3 {
				ix.byDir[k] = append(ix.byDir[k], e)
			}
		}
	}
	return ix
}

// Len is the number of titles the index knows.
func (ix *Index) Len() int {
	if ix == nil {
		return 0
	}
	return len(ix.byName)
}

// SteamName returns the manifest's title for a Steam app id.
func (ix *Index) SteamName(id int) string {
	if ix == nil {
		return ""
	}
	if e := ix.bySteam[id]; e != nil {
		return e.Name
	}
	return ""
}

// Match is who a scanned game turned out to be.
type Match struct {
	Title      string
	SteamAppID int
	GogID      string
	Confidence int    // 0–100
	How        string // how the identity was established
}

// Edition words stripped when an exact title doesn't match.

// Identify establishes the game behind a candidate. ix may be nil (no
// manifest yet): the candidate's own ids and title are used as they are.
func (ix *Index) Identify(c scan.Candidate) Match {
	m := Match{Title: c.Title, SteamAppID: c.SteamAppID, GogID: c.GogID}
	store := c.Source.Store()

	switch {
	case store && c.SteamAppID > 0:
		m.Confidence, m.How = 100, "Steam library"
		return m
	case c.SteamAppID > 0:
		m.Confidence = 95
		m.How = "Steam AppID " + strconv.Itoa(c.SteamAppID) + " read from " + c.AppIDFrom
		if n := ix.SteamName(c.SteamAppID); n != "" {
			m.Title = n
		}
		return m
	case c.GogID != "" && ix != nil && ix.byGog[c.GogID] != nil:
		e := ix.byGog[c.GogID]
		m.Confidence, m.How = 95, "GOG id "+c.GogID
		if !store || m.Title == "" {
			m.Title = e.Name
		}
		m.SteamAppID = e.SteamID
		return m
	}

	if store {
		// The store's title is authoritative; the manifest only adds a Steam id for art.
		m.Confidence, m.How = 100, c.How
		e := ix.byTitle(c.Title)
		for _, t := range []string{c.Title, scan.StripEdition(c.Title)} {
			if e == nil {
				e = ix.byLooseTitle(t)
			}
		}
		if e == nil {
			e = ix.byTitle(scan.StripEdition(c.Title))
		}
		if e != nil && e.SteamID > 0 {
			m.SteamAppID = e.SteamID
		}
		return m
	}
	if e := ix.byTitle(c.Title); e != nil {
		return Match{Title: e.Name, SteamAppID: e.SteamID, GogID: e.GogID, Confidence: 85, How: "Matched by title"}
	}
	if stripped := scan.StripEdition(c.Title); stripped != c.Title {
		if e := ix.byTitle(stripped); e != nil {
			return Match{Title: e.Name, SteamAppID: e.SteamID, GogID: e.GogID, Confidence: 75, How: "Matched by title (without edition)"}
		}
	}
	if ix != nil {
		if es := ix.byDir[scan.Normalize(filepath.Base(c.Dir))]; len(es) == 1 {
			e := es[0]
			return Match{Title: e.Name, SteamAppID: e.SteamID, GogID: e.GogID, Confidence: 70, How: "Matched by install folder name"}
		}
	}
	// Spelled a little differently, as folder names often are ("Assassin
	// Creed" for "Assassin's Creed").
	names := []string{c.Title, scan.StripEdition(c.Title), scan.ExpandAbbrev(c.Title)}
	if c.Dir != "" {
		folder := scan.CleanTitle(filepath.Base(c.Dir))
		names = append(names, folder, scan.StripEdition(folder), scan.ExpandAbbrev(folder))
	}
	for _, t := range names {
		if e := ix.byLooseTitle(t); e != nil {
			return Match{Title: e.Name, SteamAppID: e.SteamID, GogID: e.GogID, Confidence: 72, How: "Matched by a similar title"}
		}
	}
	// Unknown to the manifest: keep the name, trusting installer records more than folder names.
	m.How = "Not matched to a known game"
	if c.TitleTrusted {
		m.Confidence = 60
	} else {
		m.Confidence = 40
	}
	return m
}

func (ix *Index) byTitle(title string) *Entry {
	if ix == nil {
		return nil
	}
	k := scan.Normalize(title)
	if len(k) < 2 {
		return nil
	}
	// "Baldurs Gate 3" and "Baldur's Gate 3" normalize the same.
	return ix.byName[k]
}

// byLooseTitle finds a game whose title has the same scan.LooseKey, when
// only one game does.
func (ix *Index) byLooseTitle(title string) *Entry {
	if ix == nil {
		return nil
	}
	k := scan.LooseKey(title)
	if len(k) < 4 {
		return nil
	}
	return ix.byLoose[k]
}
