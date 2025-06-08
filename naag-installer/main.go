package main

import (
	"embed"
	"naag-installer/backend"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := backend.NewNaagInstaller()

	err := wails.Run(&options.App{
		Title:            "Naag Installer 🐍",
		Width:            700,
		Height:           600,
		Assets:           assets,
		Frameless:        false,
		DisableResize:    false,
		StartHidden:      false,
		Fullscreen:       false,
		Bind:             []interface{}{app},
		BackgroundColour: &options.RGBA{R: 15, G: 15, B: 15, A: 1},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
