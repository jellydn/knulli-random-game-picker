// Package picker owns view transitions and selection state, without SDL or firmware paths.
package picker

import (
	"encoding/json"
	"errors"
	"github.com/jellydn/knulli-random-game-picker/internal/library"
	"math/rand/v2"
	"os"
	"path/filepath"
)

type View int

const (
	Selector View = iota
	Settings
	Exit
)

type Operation int

const (
	Idle Operation = iota
	Scanning
	Verifying
)

type Filters struct {
	System    string `json:"system"`
	Favorites bool   `json:"favorites"`
}
type System struct{ ID, Label string }
type State struct {
	View          View
	Operation     Operation
	Library       library.Library
	Filters       Filters
	ExcludeRecent bool
	Recent        []string
	Focus         int
	Selected      int
	Ready         bool
	Notice        string
	RNG           *rand.Rand
}

func New(filters Filters) *State {
	return &State{Filters: filters, Selected: -1, RNG: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))}
}
func (s *State) Pool() []int {
	var recent []string
	if s.ExcludeRecent {
		recent = s.Recent
	}
	return library.Eligible(s.Library, s.Filters.System, s.Filters.Favorites, recent)
}
func (s *State) Systems() []System {
	result := []System{{"", "All Systems"}}
	seen := map[string]bool{}
	for _, game := range s.Library.Games {
		if !seen[game.System] {
			result = append(result, System{game.System, game.SystemLabel})
			seen[game.System] = true
		}
	}
	return result
}
func (s *State) Accept(lib library.Library, err error) {
	s.Operation = Idle
	if err != nil {
		s.Notice = err.Error()
		s.Ready = false
		s.Selected = -1
		return
	}
	s.Library = lib
	s.Ready = true
	s.Notice = ""
	s.Selected = -1
	found := false
	for _, system := range s.Systems() {
		if system.ID == s.Filters.System {
			found = true
		}
	}
	if !found {
		s.Filters.System = ""
	}
	if s.View == Selector && len(s.Pool()) > 0 {
		_ = s.Roll()
	}
	if len(lib.Warnings) > 0 {
		s.Notice = lib.Warnings[0]
	}
}
func (s *State) Roll() error {
	selected, err := library.Choose(s.Pool(), s.RNG)
	if err != nil {
		s.Selected = -1
		s.Notice = err.Error()
		return err
	}
	s.Selected = selected
	s.View = Selector
	s.Notice = ""
	s.Recent = append(s.Recent, s.Library.Games[selected].Path)
	if len(s.Recent) > 10 {
		s.Recent = s.Recent[len(s.Recent)-10:]
	}
	return nil
}
func (s *State) Game() *library.Game {
	if s.Selected < 0 || s.Selected >= len(s.Library.Games) {
		return nil
	}
	return &s.Library.Games[s.Selected]
}
func (s *State) Back() {
	if s.View == Selector {
		s.View = Settings
	} else {
		s.View = Exit
	}
}
func (s *State) Cycle(direction int) {
	if s.Focus == 0 {
		systems := s.Systems()
		current := 0
		for i, system := range systems {
			if system.ID == s.Filters.System {
				current = i
			}
		}
		s.Filters.System = systems[(current+direction+len(systems))%len(systems)].ID
	} else if s.Focus == 1 {
		s.Filters.Favorites = !s.Filters.Favorites
	} else if s.Focus == 2 {
		s.ExcludeRecent = !s.ExcludeRecent
	} else {
		s.Recent = nil
		s.Notice = "Recent selections cleared"
	}
}
func (s *State) ConfirmSetting() {
	if s.Focus == 0 {
		_ = s.Roll()
	} else {
		s.Cycle(1)
	}
}
func Revalidate(game library.Game, fresh library.Library) error {
	for _, candidate := range fresh.Games {
		if candidate.System == game.System && candidate.Path == game.Path {
			return nil
		}
	}
	return errors.New("Selected game is no longer launchable. Rescan from Settings")
}
func Load(dir string) Filters {
	var filters Filters
	path := filepath.Join(dir, "settings.json")
	if stat, err := os.Stat(path); err == nil && stat.Size() <= 32768 {
		if data, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(data, &filters)
		}
	}
	return filters
}
func Save(dir string, filters Filters) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(filters, "", "  ")
	if err != nil {
		return err
	}
	temp := filepath.Join(dir, "settings.json.tmp")
	if err = os.WriteFile(temp, append(data, '\n'), 0600); err != nil {
		return err
	}
	return os.Rename(temp, filepath.Join(dir, "settings.json"))
}
