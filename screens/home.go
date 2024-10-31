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

type HomeScreen struct {
	renderer      *sdl.Renderer
	listComponent *components.List
	initialized   bool
}

func NewHomeScreen(renderer *sdl.Renderer) (*HomeScreen, error) {
	return &HomeScreen{
		renderer: renderer,
		listComponent: components.NewList(
			renderer,
			18,
			sdl.Point{X: 45, Y: 95},
			func(index int, item components.Item) string {
				return fmt.Sprintf("%d. %s", index+1, item.Text)
			},
		),
	}, nil
}

func (m *HomeScreen) InitHome() {
	if m.initialized {
		return
	}
	systems := sortItemsAlphabetically(romDirsToList(listEmulatorDirs()))
	m.listComponent.SetItems(systems)
	config.CurrentSystem = systems[0].ID
	m.initialized = true
}

func romDirsToList(romDirs []RomDir) []components.Item {
	items := make([]components.Item, len(romDirs))
	for i, romDir := range romDirs {
		itemText := config.Systems[romDir.Name].Name
		if itemText == "" {
			itemText = romDir.Name
		}
		items[i] = components.Item{
			Text:  itemText,
			ID:    romDir.Name,
			Value: romDir.Path,
		}
	}
	return items
}

func sortItemsAlphabetically(items []components.Item) []components.Item {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].Text > items[j].Text {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	return items
}

func (m *HomeScreen) HandleInput(event input.InputEvent) {
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
		config.CurrentScreen = "scraping_screen"
		config.CurrentSystem = m.SelectedSystem().ID
		SetScraping()
		m.initialized = false
	}
	m.updateLogo()
}

func (m *HomeScreen) updateLogo() {
	selectedSystem := m.SelectedSystem()
	config.CurrentSystem = selectedSystem.ID
	uilib.RenderImage(m.renderer, fmt.Sprintf("%s/%s.png", config.Logos, selectedSystem.ID))
}

func (m *HomeScreen) SelectedSystem() components.Item {
	return m.listComponent.GetItems()[m.listComponent.GetSelectedIndex()]
}

func (m *HomeScreen) Draw() {
	m.InitHome()

	_ = m.renderer.SetDrawColor(0, 0, 0, 255)
	_ = m.renderer.Clear()

	uilib.RenderTexture(m.renderer, config.UiBackground, "Q2", "Q4")
	uilib.DrawText(m.renderer, "Systems", sdl.Point{X: 25, Y: 25}, config.Colors.PRIMARY, config.HeaderFont)
	uilib.RenderTexture(m.renderer, config.UiOverlaySelection, "Q2", "Q4")

	m.listComponent.Draw(config.Colors.WHITE, config.Colors.SECONDARY)

	uilib.RenderTexture(m.renderer, config.UiControls, "Q3", "Q4")
	m.updateLogo()

	m.renderer.Present()
}

type RomDir struct {
	Name string
	Path string
}

func listEmulatorDirs() []RomDir {
	romsDir := config.Roms
	dirEntries, err := os.ReadDir(romsDir)
	if err != nil {
		panic(fmt.Errorf("error reading dir %s: %w", romsDir, err))
	}

	var dirs []RomDir
	for _, entry := range dirEntries {
		if entry.IsDir() {
			dirPath := filepath.Join(romsDir, entry.Name())
			dirFiles, err := os.ReadDir(dirPath)
			if err != nil {
				panic(fmt.Errorf("error reading dir %s: %w", dirPath, err))
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
