// Package library reads an existing ES catalogue. ROMs and metadata are read-only.
package library

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Game struct {
	System, SystemLabel, Path, Title, Image string
	Favorite                                bool
}
type Library struct {
	Games    []Game
	Warnings []string
}
type Locations struct{ User, Shared, Legacy, Home string }
type Fetch func(context.Context, string) ([]byte, error)

type system struct {
	Name  string `xml:"name"`
	Label string `xml:"fullname"`
	Path  string `xml:"path"`
}
type systemList struct {
	XMLName xml.Name `xml:"systemList"`
	Systems []system `xml:"system"`
}
type metadata struct {
	Path      string `xml:"path"`
	Name      string `xml:"name"`
	Image     string `xml:"image"`
	Thumbnail string `xml:"thumbnail"`
	Marquee   string `xml:"marquee"`
}
type gamelist struct {
	Games []metadata `xml:"game"`
}
type liveRecord map[string]json.RawMessage

func (r liveRecord) text(key string) string { var s string; _ = json.Unmarshal(r[key], &s); return s }
func (r liveRecord) flag(key string) bool   { return string(r[key]) == "true" || r.text(key) == "true" }

func readXML(path string, value any) error {
	data, err := readLimited(path, 32<<20)
	if err != nil {
		return err
	}
	if err = xml.Unmarshal(data, value); err != nil {
		return fmt.Errorf("Invalid XML %s: %w", path, err)
	}
	return nil
}
func readLimited(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if stat.Size() > limit {
		return nil, fmt.Errorf("File exceeds limit: %s", path)
	}
	// Limit the read as well as stat, including when a file grows during discovery.
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("File exceeds %d byte limit", limit)
	}
	return data, err
}

func resolve(path, root, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(root, path)
}
func identity(path string) string {
	if canonical, err := filepath.EvalSymlinks(path); err == nil {
		return canonical
	}
	return filepath.Clean(path)
}
func exists(path string) bool { _, err := os.Stat(path); return err == nil }

func Discover(ctx context.Context, loc Locations, fetch Fetch) (Library, error) {
	var result Library
	var systems systemList
	base := ""
	for _, path := range []string{filepath.Join(loc.User, "es_systems_custom.cfg"), filepath.Join(loc.User, "es_systems.cfg"), filepath.Join(loc.Shared, "es_systems.cfg"), filepath.Join(loc.Legacy, "es_systems.cfg")} {
		if exists(path) {
			base = path
			break
		}
	}
	if base == "" {
		return result, errors.New("No EmulationStation system configuration found")
	}
	if err := readXML(base, &systems); err != nil {
		return result, err
	}
	for _, dir := range []string{loc.User, loc.Shared} {
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasPrefix(name, "es_systems_") || !strings.HasSuffix(name, ".cfg") {
				continue
			}
			var extra systemList
			if err := readXML(filepath.Join(dir, name), &extra); err != nil {
				return result, err
			}
			for _, addition := range extra.Systems {
				if addition.Name == "" {
					continue
				}
				found := false
				for i := range systems.Systems {
					if systems.Systems[i].Name != addition.Name {
						continue
					}
					if addition.Label != "" {
						systems.Systems[i].Label = addition.Label
					}
					if addition.Path != "" {
						systems.Systems[i].Path = addition.Path
					}
					found = true
					break
				}
				if !found {
					systems.Systems = append(systems.Systems, addition)
				}
			}
		}
	}
	get := func(route string) ([]liveRecord, error) {
		body, err := fetch(ctx, route)
		if err != nil {
			return nil, err
		}
		var records []liveRecord
		if err := json.Unmarshal(body, &records); err != nil || strings.TrimSpace(string(body)) == "null" {
			return nil, fmt.Errorf("Unsupported ES response for %s", route)
		}
		return records, nil
	}
	liveSystems, err := get("/systems")
	if err != nil {
		return result, err
	}
	counts := map[string]int{}
	executable, _ := os.Executable()
	ownLauncher := identity(filepath.Join(filepath.Dir(executable), "RandomGamePicker.sh"))
	for _, liveSystem := range liveSystems {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if liveSystem.flag("collection") || !liveSystem.flag("gamesystem") {
			continue
		}
		id := liveSystem.text("name")
		if id == "" {
			continue
		}
		games, err := get("/systems/" + url.PathEscape(id) + "/games")
		if err != nil {
			return result, err
		}
		var config system
		for _, candidate := range systems.Systems {
			if candidate.Name == id {
				config = candidate
				break
			}
		}
		root := resolve(config.Path, loc.Home, loc.Home)
		supported := config.Path != "" && filepath.IsAbs(root) && id != "ports"
		label := config.Label
		if label == "" {
			label = id
		}
		meta := map[string]metadata{}
		if supported {
			path := filepath.Join(root, "gamelist.xml")
			if !exists(path) {
				path = filepath.Join(loc.User, "gamelists", id, "gamelist.xml")
			}
			if exists(path) {
				var list gamelist
				if err := readXML(path, &list); err != nil {
					result.Warnings = append(result.Warnings, err.Error())
				} else {
					for _, game := range list.Games {
						meta[identity(resolve(game.Path, root, loc.Home))] = game
					}
				}
			}
		}
		for _, liveGame := range games {
			path := liveGame.text("path")
			if path == "" || len(path) > 4096 || strings.ContainsRune(path, 0) {
				continue
			}
			key := identity(path)
			counts[key]++ // Count unsupported/hidden systems too: ES launch searches all real systems.
			if !supported || key == ownLauncher || !filepath.IsAbs(path) || !exists(path) || liveGame.flag("hidden") {
				continue
			}
			node := meta[key]
			game := Game{System: id, SystemLabel: label, Path: path, Title: liveGame.text("name"), Favorite: liveGame.flag("favorite")}
			if game.Title == "" {
				game.Title = node.Name
			}
			if game.Title == "" {
				game.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			}
			// Bound untrusted UI metadata and keep valid UTF-8 at the boundary.
			if len(game.Title) > 4096 {
				game.Title = strings.ToValidUTF8(game.Title[:4096], "")
			}
			for _, art := range []string{node.Image, node.Thumbnail, node.Marquee} {
				if art != "" && exists(resolve(art, root, loc.Home)) {
					game.Image = resolve(art, root, loc.Home)
					break
				}
			}
			result.Games = append(result.Games, game)
		}
	}
	unique := result.Games[:0]
	for _, game := range result.Games {
		if counts[identity(game.Path)] == 1 {
			unique = append(unique, game)
		}
	}
	if len(unique) != len(result.Games) {
		result.Warnings = append(result.Warnings, "Ambiguous paths excluded to prevent multiple launches")
	}
	result.Games = unique
	return result, nil
}

func Eligible(lib Library, system string, favorites bool, recent []string) []int {
	excluded := map[string]bool{}
	for _, path := range recent {
		excluded[path] = true
	}
	var pool []int
	for i, game := range lib.Games {
		if (system != "" && game.System != system) || (favorites && !game.Favorite) || excluded[game.Path] {
			continue
		}
		pool = append(pool, i)
	}
	return pool
}
func Choose(pool []int, rng *rand.Rand) (int, error) {
	if len(pool) == 0 {
		return 0, errors.New("No eligible games. Change filters or reset recent selections")
	}
	return pool[rng.IntN(len(pool))], nil
}
