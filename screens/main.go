package screens

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anibaldeboni/screech/components"
	"github.com/anibaldeboni/screech/config"
	"github.com/anibaldeboni/screech/input"
	"github.com/anibaldeboni/screech/uilib"

	"github.com/veandco/go-sdl2/sdl"
)

type MainScreen struct {
	renderer      *sdl.Renderer
	listComponent *components.ListComponent
	initialized   bool
}

func NewMainScreen(renderer *sdl.Renderer) (*MainScreen, error) {
	listComponent := components.NewListComponent(renderer, 19, func(index int, item components.Item) string {
		return fmt.Sprintf("%d. %s", index+1, item.Text)
	})
	return &MainScreen{
		renderer:      renderer,
		listComponent: listComponent,
	}, nil
}

func (m *MainScreen) InitMain() {
	if m.initialized {
		return
	}

	m.listComponent.SetItems(romDirsToList(listEmulatorDirs()))
	m.initialized = true
}

func romDirsToList(romDirs []RomDir) []components.Item {
	var items []components.Item
	for _, romDir := range romDirs {
		items = append(items, components.Item{
			Text:  romDir.Name,
			Value: romDir.Path,
		})
	}
	return items
}

func (m *MainScreen) HandleInput(event input.InputEvent) {
	switch event.KeyCode {
	case "DOWN":
		m.listComponent.ScrollDown()
	case "UP":
		m.listComponent.ScrollUp()
	case "B":
		os.Exit(0)
	case "A":
		if len(m.listComponent.GetItems()) == 0 {
			return
		}
		selectedSystem := m.listComponent.GetItems()[m.listComponent.GetSelectedIndex()]
		config.CurrentSystem = selectedSystem.Text
		config.CurrentScreen = "scraping_screen"
		m.initialized = false
	}
}

func (m *MainScreen) Draw() {
	m.InitMain()

	m.renderer.SetDrawColor(0, 0, 0, 255) // Background color
	m.renderer.Clear()

	uilib.RenderTexture(m.renderer, config.UiBackground, "Q2", "Q4")

	// Draw the title
	uilib.DrawText(m.renderer, "Systems", sdl.Point{X: 25, Y: 25}, config.Colors.PRIMARY, config.HeaderFont)

	m.listComponent.Draw(config.Colors.WHITE, config.Colors.SECONDARY)

	uilib.RenderTexture(m.renderer, config.UiControls, "Q3", "Q4")

	m.renderer.Present()
}

type RomDir struct {
	Name string
	Path string
}

func listEmulatorDirs() []RomDir {
	emulatorsDir := config.Roms
	dirEntries, err := os.ReadDir(emulatorsDir)
	if err != nil {
		panic(err)
	}

	var dirs []RomDir
	for _, entry := range dirEntries {
		if entry.IsDir() {
			dirPath := filepath.Join(emulatorsDir, entry.Name())
			dirFiles, err := os.ReadDir(dirPath)
			if err != nil {
				panic(err)
			}
			if len(dirFiles) > 0 {
				dirs = append(dirs, RomDir{
					Name: entry.Name(),
					Path: dirPath,
				})
			}
		}
	}

	return dirs
}
