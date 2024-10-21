package config

import (
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

var (
	configFile         = "screech.yaml"
	Debug              bool
	ScreenWidth        = int32(1280)
	ScreenHeight       = int32(720)
	CurrentScreen      string
	CurrentSystem      string
	CurrentGame        string
	BodyFont           *ttf.Font
	HeaderFont         *ttf.Font
	ListFont           *ttf.Font
	LongTextFont       *ttf.Font
	Colors             FontColors
	ControlType        string
	Roms               string
	Logos              string
	UiControls         = "assets/ui_controls_1280_720.bmp"
	UiBackground       = "assets/bg.bmp"
	UiOverlay          = "assets/bg_overlay.bmp"
	UiOverlaySelection = "assets/bg_overlay_selection.bmp"
	Username           string
	Password           string
	SystemsIDs         map[string]string
	SystemsNames       map[string]string
	GameRegions        []string
	Media              ScrapeMedia
	Thumbnail          = thumbConfig{
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
)

func InitVars() {
	cfg, err := readConfigFile()
	if err != nil {
		SaveCurrent()
		return
	}
	Debug = cfg.Debug
	CurrentScreen = "main_screen"
	CurrentSystem = ""
	CurrentGame = ""
	ControlType = "keyboard"
	Roms = cfg.Roms
	Logos = cfg.Logos
	MaxScanDepth = cfg.MaxScanDepth
	if len(cfg.ExcludeExtensions) == 0 {
		ExcludeExtensions = defaultExcludeExtensions
	} else {
		ExcludeExtensions = cfg.ExcludeExtensions
	}
	Username = cfg.Screenscraper.Username
	Password = cfg.Screenscraper.Password
	Threads = cfg.Screenscraper.Threads
	SystemsIDs = defineSystemsIDs(cfg.Screenscraper.Systems)
	SystemsNames = defineSystemsNames(cfg.Screenscraper.Systems)
	GameRegions = cfg.Screenscraper.Media.Regions
	Media = cfg.Screenscraper.Media
	Thumbnail = cfg.Thumbnail
	BodyFont = nil
	HeaderFont = nil
	ListFont = nil
	LongTextFont = nil
	Colors = FontColors{
		WHITE: sdl.Color{R: 255, G: 255, B: 255, A: 255},
		// PRIMARY:   sdl.Color{R: 255, G: 214, B: 255, A: 255},
		PRIMARY:   sdl.Color{R: 113, G: 255, B: 142, A: 255},
		SECONDARY: sdl.Color{R: 231, G: 192, B: 255, A: 255},
		BLACK:     sdl.Color{R: 0, G: 0, B: 0, A: 255},
	}
}

func defineSystemsIDs(systems []scraperSystem) map[string]string {
	systemsIDs := make(map[string]string, len(systems))
	for _, system := range systems {
		systemsIDs[system.Dir] = system.ID
	}
	return systemsIDs
}
func defineSystemsNames(systems []scraperSystem) map[string]string {
	systemsNames := make(map[string]string, len(systems))
	for _, system := range systems {
		systemsNames[system.Dir] = system.Name
	}
	return systemsNames
}
func ScrapedImgDir() string {
	dir := strings.ReplaceAll(Thumbnail.Dir, "/", string(filepath.Separator))
	dir = strings.ReplaceAll(dir, "\\", string(filepath.Separator))
	dir = strings.ReplaceAll(dir, "%SYSTEM%", CurrentSystem)
	return dir
}

type scraperConfig struct {
	Username string          `yaml:"username"`
	Password string          `yaml:"password"`
	Media    ScrapeMedia     `yaml:"media"`
	Systems  []scraperSystem `yaml:"systems"`
	Threads  int             `yaml:"threads"`
}

type scraperSystem struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
	Dir  string `yaml:"dir"`
}

type ScrapeMedia struct {
	Type    string   `yaml:"type"`
	Regions []string `yaml:"regions"`
	Width   int32    `yaml:"width"`
	Height  int32    `yaml:"height"`
}

type thumbConfig struct {
	Dir    string `yaml:"dir"`
	Width  int    `yaml:"width"`
	Height int    `yaml:"height"`
}
type userConfigs struct {
	Thumbnail         thumbConfig   `yaml:"thumbnail"`
	Roms              string        `yaml:"roms"`
	Logos             string        `yaml:"logos"`
	Screenscraper     scraperConfig `yaml:"screenscraper"`
	MaxScanDepth      int           `yaml:"max-scan-depth"`
	ExcludeExtensions []string      `yaml:"exclude-extensions"`
	Debug             bool          `yaml:"debug,omitempty"`
}

func readConfigFile() (*userConfigs, error) {
	var cfg *userConfigs
	file, err := os.ReadFile(configFile)

	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, &cfg)
	return cfg, err
}

func SaveCurrent() {
	config := userConfigs{
		Roms: Roms,
		Screenscraper: scraperConfig{
			Username: Username,
			Password: Password,
			Media:    Media,
		},
		Thumbnail: Thumbnail,
		Debug:     Debug,
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		panic(err)
	}
}
