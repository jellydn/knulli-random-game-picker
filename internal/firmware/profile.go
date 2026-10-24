// Package firmware describes device policy, separate from library and launcher backends.
package firmware

import (
	"fmt"
	"github.com/jellydn/knulli-random-game-picker/internal/library"
	"os"
)

type Profile struct {
	ID, Label, StateDir, ControllerDB string
	Locations                         library.Locations
	SwapFaceButtons                   bool
	CanLaunch                         bool
}

func Resolve(id string) (Profile, error) {
	if id != "knulli" {
		return Profile{}, fmt.Errorf("Unsupported firmware %q; this build supports KNULLI only", id)
	}
	home := os.Getenv("HOME")
	if home == "" {
		home = "/userdata/system"
	}
	state := os.Getenv("PICKER_STATE_DIR")
	if state == "" {
		state = "/userdata/system/configs/random-game-picker"
	}
	return Profile{ID: id, Label: "KNULLI", StateDir: state, ControllerDB: "/tmp/gamecontrollerdb.txt", CanLaunch: true,
		SwapFaceButtons: os.Getenv("PICKER_CONTROLLER_LAYOUT") == "knulli",
		Locations:       library.Locations{User: "/userdata/system/configs/emulationstation", Shared: "/usr/share/emulationstation", Legacy: "/etc/emulationstation", Home: home}}, nil
}
