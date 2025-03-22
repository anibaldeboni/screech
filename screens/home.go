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

	"slices"

	"github.com/veandco/go-sdl2/sdl"
)

type HomeScreen struct {
	renderer    *sdl.Renderer
	systemsList *components.List[romDirSettings]
	textView    *components.TextView
	initialized bool
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
		systemsList: components.NewList(
			renderer,
			18,
			sdl.Point{X: 45, Y: 95},
			func(index int, item components.Item[romDirSettings]) string {
				return fmt.Sprintf("%d. %s", index+1, item.Label)
			},
		),
		textView: components.NewTextView(
			renderer,
			components.TextViewSize{Width: 50, Height: 18},
			sdl.Point{X: 545, Y: 96},
		),
	}, nil
}

func (h *HomeScreen) InitHome() {
	if h.initialized {
		return
	}

	romDirs, err := listRomsDirs()
	if err != nil {
		h.textView.AddText(err.Error())
	}
	systems := sortItemsAlphabetically(romDirsToList(romDirs))
	h.systemsList.SetItems(systems)
	h.initialized = true
}

func (h *HomeScreen) HandleInput(event input.UserInputEvent) {
	switch event.KeyCode {
	case "DOWN":
		h.systemsList.ScrollDown()
	case "UP":
		h.systemsList.ScrollUp()
	case "B":
		os.Exit(0)
	case "A":
		if h.isNotInErrorMode() {
			h.goToScraping([]romDirSettings{h.systemsList.SelectedValue()})
		}
	case "X":
		if h.isNotInErrorMode() {
			h.goToScraping(h.systemsList.GetValues())
		}
	}
}

func (h *HomeScreen) isNotInErrorMode() bool {
	return len(h.textView.GetText()) == 0
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
	selectedSystem := h.systemsList.SelectedValue()
	logoPath := fmt.Sprintf("%s/%s.png", config.LogosBaseDir, selectedSystem.DirName)

	if _, err := os.Stat(logoPath); err == nil {
		uilib.RenderImage(h.renderer, logoPath)
	}
}

func (h *HomeScreen) Draw() {
	h.InitHome()

	_ = h.renderer.SetDrawColor(0, 0, 0, 255)
	_ = h.renderer.Clear()

	uilib.RenderTexture(h.renderer, config.UiBackground, "Q2", "Q4")
	uilib.DrawText(h.renderer, "SCREECH", sdl.Point{X: 25, Y: 25}, config.Colors.PRIMARY, config.HeaderFont)
	uilib.DrawText(h.renderer, config.Version, sdl.Point{X: 1200, Y: 35}, config.Colors.PRIMARY, config.LongTextFont)
	uilib.RenderTexture(h.renderer, config.UiOverlaySelection, "Q2", "Q4")

	h.systemsList.Draw(config.Colors.WHITE, config.Colors.SECONDARY)

	uilib.RenderTexture(h.renderer, config.UiControls, "Q3", "Q4")

	if len(h.textView.GetText()) > 0 {
		h.textView.Draw(config.Colors.WHITE)
	} else {
		h.updateLogo()
	}

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
	slices.SortFunc(items, func(a, b components.Item[romDirSettings]) int {
		if a.Label < b.Label {
			return -1
		}
		if a.Label > b.Label {
			return 1
		}
		return 0
	})
	return items
}

type RomDir struct {
	Name string
	Path string
}

func listRomsDirs() ([]RomDir, error) {
	dirEntries, err := os.ReadDir(config.RomsBaseDir)
	if err != nil {
		return nil, fmt.Errorf("error reading dir %s: %w", config.RomsBaseDir, err)
	}

	var dirs []RomDir
	for _, entry := range filterDirs(dirEntries) {
		if isAllowedDir(entry.Name()) {
			dirPath := filepath.Join(config.RomsBaseDir, entry.Name())
			subDirEntries, err := os.ReadDir(dirPath)
			if err != nil {
				return nil, fmt.Errorf("error reading dir %s: %w", dirPath, err)
			}
			if len(subDirEntries) > 0 && hasVisibleEntries(subDirEntries) {
				dirs = append(dirs, RomDir{
					Name: entry.Name(),
					Path: dirPath,
				})
			}
		}
	}

	return dirs, nil
}

func filterDirs(entries []os.DirEntry) []os.DirEntry {
	var dirs []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			dirs = append(dirs, entry)
		}
	}
	return dirs
}

func hasVisibleEntries(files []os.DirEntry) bool {
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") {
			return true
		}
	}
	return false
}

func isAllowedDir(dirName string) bool {
	return !slices.Contains(config.IgnoreDirs, dirName)
}
