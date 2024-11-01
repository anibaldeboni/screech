package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	yaml "gopkg.in/yaml.v3"
)

type FontColors struct {
	WHITE     sdl.Color
	PRIMARY   sdl.Color
	SECONDARY sdl.Color
	BLACK     sdl.Color
}

type SystemSettings struct {
	ID        string `yaml:"id"`
	Name      string `yaml:"name"`
	OutputDir string `yaml:"output-dir,omitempty"`
}

type ScrapeMedia struct {
	Type                string   `yaml:"type"`
	Regions             []string `yaml:"regions"`
	Width               int32    `yaml:"width"`
	Height              int32    `yaml:"height"`
	IgnoreMissingRegion bool     `yaml:"ignore-missing-region"`
}

type scraperConfig struct {
	Username string      `yaml:"username"`
	Password string      `yaml:"password"`
	Media    ScrapeMedia `yaml:"media"`
	Threads  int         `yaml:"threads"`
}

type scraperSystem struct {
	ID        string `yaml:"id"`
	Name      string `yaml:"name"`
	OutputDir string `yaml:"output-dir,omitempty"`
	Dir       string `yaml:"dir"`
}

type boxartConfig struct {
	Dir    string `yaml:"dir"`
	Width  int    `yaml:"width"`
	Height int    `yaml:"height"`
}
type userConfigs struct {
	Boxart                  boxartConfig    `yaml:"thumbnail"`
	Roms                    string          `yaml:"roms"`
	Logos                   string          `yaml:"logos"`
	Screenscraper           scraperConfig   `yaml:"screenscraper"`
	Systems                 []scraperSystem `yaml:"systems"`
	MaxScanDepth            int             `yaml:"max-scan-depth"`
	ExcludeExtensions       []string        `yaml:"exclude-extensions"`
	IgnoreSkippedRomMessage bool            `yaml:"ignore-skipped-rom-message,omitempty"`
	Debug                   bool            `yaml:"debug,omitempty"`
}

var (
	Version            = "dev"
	ConfigFile         = "screech.yaml"
	Debug              bool
	ScreenWidth        = int32(1280)
	ScreenHeight       = int32(720)
	CurrentScreen      string
	BodyFont           *ttf.Font
	HeaderFont         *ttf.Font
	ListFont           *ttf.Font
	LongTextFont       *ttf.Font
	Colors             FontColors
	RomsBaseDir        string
	LogosBaseDir       string
	UiControls         = "assets/ui_controls_1280_720.bmp"
	UiBackground       = "assets/bg.bmp"
	UiOverlay          = "assets/bg_overlay.bmp"
	UiOverlaySelection = "assets/bg_overlay_selection.bmp"
	Username           string
	Password           string
	Systems            map[string]SystemSettings
	Media              ScrapeMedia
	Boxart             = boxartConfig{
		Width:  400,
		Height: 580,
		Dir:    "thumbnails",
	}
	Threads                  = 1
	MaxScanDepth             = 2
	ExcludeExtensions        []string
	defaultExcludeExtensions = []string{
		".cue",
		".m3u",
		".jpg",
		".png",
		".sub",
		".db",
		".xml",
		".txt",
		".dat",
		".miyoocmd",
		".cfg",
		".state",
		".srm",
	}
	IgnoreSkippedRomMessage bool
)

func InitVars() {
	cfg, err := readConfigFile()
	if err != nil {
		SaveCurrent()
		return
	}
	Debug = cfg.Debug
	CurrentScreen = "home_screen"
	RomsBaseDir = cfg.Roms
	LogosBaseDir = cfg.Logos
	MaxScanDepth = cfg.MaxScanDepth
	if len(cfg.ExcludeExtensions) == 0 {
		ExcludeExtensions = defaultExcludeExtensions
	} else {
		ExcludeExtensions = cfg.ExcludeExtensions
	}
	IgnoreSkippedRomMessage = cfg.IgnoreSkippedRomMessage
	Username = cfg.Screenscraper.Username
	Password = cfg.Screenscraper.Password
	Threads = cfg.Screenscraper.Threads
	Systems = setSystems(cfg.Systems)
	Media = cfg.Screenscraper.Media
	Boxart = cfg.Boxart
	BodyFont = nil
	HeaderFont = nil
	ListFont = nil
	LongTextFont = nil
	Colors = FontColors{
		WHITE:     sdl.Color{R: 255, G: 255, B: 255, A: 255},
		PRIMARY:   sdl.Color{R: 113, G: 255, B: 142, A: 255},
		SECONDARY: sdl.Color{R: 168, G: 48, B: 190, A: 255},
		BLACK:     sdl.Color{R: 0, G: 0, B: 0, A: 255},
	}
}

func setSystems(systems []scraperSystem) map[string]SystemSettings {
	systemSettings := make(map[string]SystemSettings)
	for _, system := range systems {
		if system.Dir == "" {
			fmt.Printf("Skipping system with empty Dir: %+v\n", system)
			continue
		}

		var outputDir string
		if system.OutputDir != "" {
			outputDir = system.OutputDir
		} else {
			outputDir = system.Dir
		}

		systemSettings[system.Dir] = SystemSettings{
			ID:        system.ID,
			Name:      system.Name,
			OutputDir: outputDir,
		}
	}

	return systemSettings
}

func ScrapedImgDir(outputDir string) string {
	dir := strings.ReplaceAll(Boxart.Dir, "/", string(filepath.Separator))
	dir = strings.ReplaceAll(dir, "\\", string(filepath.Separator))
	dir = strings.ReplaceAll(dir, "%SYSTEM%", outputDir)
	return dir
}

func readConfigFile() (*userConfigs, error) {
	var cfg *userConfigs
	file, err := os.ReadFile(ConfigFile)

	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, &cfg)
	return cfg, err
}

func SaveCurrent() {
	config := userConfigs{
		Roms: RomsBaseDir,
		Screenscraper: scraperConfig{
			Username: Username,
			Password: Password,
			Media:    Media,
		},
		Boxart: Boxart,
		Debug:  Debug,
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(ConfigFile, data, 0644); err != nil {
		panic(err)
	}
}
