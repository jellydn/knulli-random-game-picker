package main

import (
	"context"
	"fmt"
	"github.com/jellydn/knulli-random-game-picker/internal/launcher"
	"os"
)

func main() {
	if len(os.Args) != 4 {
		os.Exit(2)
	}
	body, err := launcher.Request(context.Background(), os.Args[1], os.Args[2], os.Args[3])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(string(body))
}
