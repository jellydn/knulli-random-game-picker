package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/jellydn/knulli-random-game-picker/internal/firmware"
	"github.com/jellydn/knulli-random-game-picker/internal/launcher"
	"github.com/jellydn/knulli-random-game-picker/internal/library"
	"github.com/jellydn/knulli-random-game-picker/internal/picker"
	"github.com/jellydn/knulli-random-game-picker/internal/ui"
)

func main() {
	runtime.LockOSThread()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Random Game Picker:", err)
		os.Exit(1)
	}
}
func run() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	assets := flag.String("assets", filepath.Join(filepath.Dir(executable), "assets"), "Bundled font directory")
	demo := flag.Bool("demo", false, "Synthetic library, no launch or saved settings")
	windowed := flag.Bool("windowed", false, "Desktop window")
	size := flag.String("size", "1280x720", "Window WIDTHxHEIGHT")
	config := flag.String("es-config", "", "ES user configuration directory")
	screenshot := flag.String("screenshot", "", "Save first ready frame as PNG")
	version := flag.Bool("version", false, "Print version")
	firmwareID := flag.String("firmware", "knulli", "Firmware profile (KNULLI only)")
	flag.Parse()
	if *version {
		fmt.Println("0.2.2")
		return nil
	}
	profile, err := firmware.Resolve(*firmwareID)
	if err != nil {
		return err
	}
	if *config != "" {
		profile.Locations.User = *config
	}
	dimensions := strings.Split(*size, "x")
	if len(dimensions) != 2 {
		return fmt.Errorf("Expected WIDTHxHEIGHT")
	}
	width, e1 := strconv.Atoi(dimensions[0])
	height, e2 := strconv.Atoi(dimensions[1])
	if e1 != nil || e2 != nil || width < 640 || height < 480 || width > 3840 || height > 2160 {
		return fmt.Errorf("Unsupported size")
	}
	fmt.Printf("Firmware: %s; KNULLI face-button correction: %t\n", profile.Label, profile.SwapFaceButtons)
	filters := picker.Filters{}
	if !*demo {
		filters = picker.Load(profile.StateDir)
	}
	state := picker.New(filters)
	if *demo {
		state.Accept(library.Library{Games: []library.Game{
			{System: "snes", SystemLabel: "Super Nintendo", Path: "demo:snes", Title: "A new adventure awaits", Favorite: true},
			{System: "gba", SystemLabel: "Game Boy Advance", Path: "demo:gba", Title: "The pocket-sized quest"},
		}}, nil)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scan := func(ctx context.Context) (library.Library, error) {
		return library.Discover(ctx, profile.Locations, launcher.Fetch)
	}
	game, err := ui.Run(ctx, ui.Options{Profile: profile, Demo: *demo, Windowed: *windowed, Width: width, Height: height, Assets: *assets, Screenshot: *screenshot}, state, scan)
	if err != nil {
		return err
	}
	cancel() // Stop a pending scan before saving/exiting; never cancel the handoff context.
	if !*demo {
		if err := picker.Save(profile.StateDir, state.Filters); err != nil {
			return err
		}
	}
	if game != nil {
		fmt.Println("SDL closed. Submitting one exact-path ES launch request; acceptance is not proof of play.")
		return launcher.Submit(context.Background(), game.Path)
	}
	return nil
}
