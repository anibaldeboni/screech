package config

import (
	"os"
	"path/filepath"

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
	Debug           bool
	ScreenWidth     int32
	ScreenHeight    int32
	CurrentPlatform string
	CurrentScreen   string
	CurrentSystem   string
	CurrentGame     string
	CurrentTester   string
	BodyFont        *ttf.Font
	HeaderFont      *ttf.Font
	BodyBigFont     *ttf.Font
	LongTextFont    *ttf.Font
	Colors          FontColors
	ControlType     string
	Roms            string
	UiControls      = "assets/ui_controls_1280_720.bmp"
	UiBackground    = "assets/bg.bmp"
	UiOverlay       = "assets/bg_overlay.bmp"
	Username        string
	Password        string
	SystemsIDs      map[string]string
	GameRegions     []string
)

func InitVars() {
	config, err := readConfigFile()
	if err != nil {
		panic(err)
	}
	Debug = config.Debug
	ScreenWidth = 0
	ScreenHeight = 0
	CurrentPlatform = "tsp"
	CurrentScreen = "main_screen"
	CurrentSystem = ""
	CurrentGame = ""
	ControlType = "keyboard"
	ScreenWidth = config.Screen.Width
	ScreenHeight = config.Screen.Height
	Roms = config.Roms
	Username = config.Screenscraper.Username
	Password = config.Screenscraper.Password
	SystemsIDs = config.Screenscraper.SystemsIDs
	GameRegions = config.Screenscraper.Media.Regions
	BodyFont = nil
	HeaderFont = nil
	BodyBigFont = nil
	LongTextFont = nil
	Colors = FontColors{
		WHITE:     sdl.Color{R: 255, G: 255, B: 255, A: 255},
		PRIMARY:   sdl.Color{R: 255, G: 214, B: 255, A: 255},
		SECONDARY: sdl.Color{R: 231, G: 192, B: 255, A: 255},
		BLACK:     sdl.Color{R: 0, G: 0, B: 0, A: 255},
	}
}

func ScrapedImgDir() string {
	return filepath.Join(Roms, CurrentSystem, "Imgs")
}

type scraperConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Media    struct {
		Type    string   `yaml:"type"`
		Width   int32    `yaml:"width"`
		Height  int32    `yaml:"height"`
		Regions []string `yaml:"regions"`
	} `yaml:"media"`
	SystemsIDs map[string]string `yaml:"systems-ids"`
}

type userConfigs struct {
	Screen struct {
		Width  int32 `yaml:"width"`
		Height int32 `yaml:"height"`
	} `yaml:"screen"`
	Roms          string        `yaml:"roms"`
	Screenscraper scraperConfig `yaml:"screenscraper"`
	Debug         bool          `yaml:"debug,omitempty"`
}

func readConfigFile() (*userConfigs, error) {
	configFile := "./screech.yaml"
	file, err := os.ReadFile(configFile)
	if err != nil {
		return &userConfigs{}, err
	}
	var config *userConfigs
	err = yaml.Unmarshal(file, &config)
	return config, err
}
