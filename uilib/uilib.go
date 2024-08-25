package uilib

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func InitSDL() error {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_JOYSTICK | sdl.INIT_GAMECONTROLLER); err != nil {
		return fmt.Errorf("error initializing SDL: %w", err)
	}
	return nil
}

func InitTTF() error {
	if err := ttf.Init(); err != nil {
		return fmt.Errorf("error initializing SDL_ttf: %w", err)
	}
	return nil
}

func InitFont(fontTtf []byte, font **ttf.Font, size int) error {
	rwops, err := sdl.RWFromMem(fontTtf)
	if err != nil {
		return fmt.Errorf("error creating RWops from memory: %w", err)
	}
	f, err := ttf.OpenFontRW(rwops, 1, size)
	if err != nil {
		return fmt.Errorf("error loading font: %w", err)
	}
	*font = f
	return nil
}
