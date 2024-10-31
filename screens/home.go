package screens

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anibaldeboni/screech/components"
	"github.com/anibaldeboni/screech/config"
	"github.com/anibaldeboni/screech/input"
	"github.com/anibaldeboni/screech/uilib"

	"github.com/veandco/go-sdl2/sdl"
)

type HomeScreen struct {
	renderer      *sdl.Renderer
	listComponent *components.List[romDirSettings]
	initialized   bool
}

type romDirSettings struct {
	DirName    string
	Path       string
	OutputDir  string
	SystemName string
	SystemID   string
}

func NewHomeScreen(renderer *sdl.Renderer) (*HomeScreen, error) {
	return &HomeScreen{
		renderer: renderer,
		listComponent: components.NewList(
			renderer,
			18,
			sdl.Point{X: 45, Y: 95},
			func(index int, item components.Item[romDirSettings]) string {
				return fmt.Sprintf("%d. %s", index+1, item.Label)
			},
		),
	}, nil
}

func (m *HomeScreen) InitHome() {
	if m.initialized {
		return
	}
	systems := sortItemsAlphabetically(romDirsToList(listRomsDirs()))
	m.listComponent.SetItems(systems)
	m.initialized = true
}

func (h *HomeScreen) HandleInput(event input.InputEvent) {
	switch event.KeyCode {
	case "DOWN":
		h.listComponent.ScrollDown()
	case "UP":
		h.listComponent.ScrollUp()
	case "B":
		os.Exit(0)
	case "A":
		h.goToScraping([]romDirSettings{h.listComponent.SelectedValue()})
	case "X":
		h.goToScraping(h.listComponent.GetValues())
	}
	h.updateLogo()
}

func (h *HomeScreen) goToScraping(systems []romDirSettings) {
	if len(systems) == 0 {
		return
	}
	config.CurrentScreen = "scraping_screen"
	SetTargetSystems(systems)
	SetScraping()
	h.initialized = false
}

func (h *HomeScreen) updateLogo() {
	selectedSystem := h.listComponent.SelectedValue()
	uilib.RenderImage(h.renderer, fmt.Sprintf("%s/%s.png", config.LogosBaseDir, selectedSystem.DirName))
}

func (h *HomeScreen) Draw() {
	h.InitHome()

	_ = h.renderer.SetDrawColor(0, 0, 0, 255)
	_ = h.renderer.Clear()

	uilib.RenderTexture(h.renderer, config.UiBackground, "Q2", "Q4")
	uilib.DrawText(h.renderer, "Systems", sdl.Point{X: 25, Y: 25}, config.Colors.PRIMARY, config.HeaderFont)
	uilib.RenderTexture(h.renderer, config.UiOverlaySelection, "Q2", "Q4")

	h.listComponent.Draw(config.Colors.WHITE, config.Colors.SECONDARY)

	uilib.RenderTexture(h.renderer, config.UiControls, "Q3", "Q4")
	h.updateLogo()

	h.renderer.Present()
}

func romDirsToList(romDirs []RomDir) []components.Item[romDirSettings] {
	items := make([]components.Item[romDirSettings], 0, len(romDirs))
	for _, romDir := range romDirs {
		system := config.Systems[romDir.Name]
		label := system.Name
		if label == "" {
			label = romDir.Name
		}
		items = append(items, components.Item[romDirSettings]{
			Label: label,
			Value: romDirSettings{
				DirName:    romDir.Name,
				Path:       romDir.Path,
				OutputDir:  system.OutputDir,
				SystemName: label,
				SystemID:   system.ID,
			},
		})
	}
	return items
}

func sortItemsAlphabetically(items []components.Item[romDirSettings]) []components.Item[romDirSettings] {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].Label > items[j].Label {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	return items
}

type RomDir struct {
	Name string
	Path string
}

func listRomsDirs() []RomDir {
	romsDir := config.RomsBaseDir
	dirEntries, err := os.ReadDir(romsDir)
	if err != nil {
		panic(fmt.Errorf("error reading dir %s: %w", romsDir, err))
	}

	var dirs []RomDir
	for _, entry := range dirEntries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
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
