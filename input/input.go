package input

import "github.com/veandco/go-sdl2/sdl"

type InputEvent struct {
	KeyCode string
}

var InputChannel = make(chan InputEvent)

func StartListening() {
	go listenForKeyboardEvents()
	go listenForControllerEvents()
}

func listenForKeyboardEvents() {
	previousKeyState := make([]uint8, sdl.NUM_SCANCODES)

	keyMappings := map[sdl.Scancode]string{
		sdl.SCANCODE_DOWN:   "DOWN",
		sdl.SCANCODE_UP:     "UP",
		sdl.SCANCODE_A:      "A",
		sdl.SCANCODE_B:      "B",
		sdl.SCANCODE_RETURN: "START",
		sdl.SCANCODE_ESCAPE: "SELECT",
	}

	for {
		currentKeyState := sdl.GetKeyboardState()

		for scancode, keyCode := range keyMappings {
			if currentKeyState[scancode] != 0 && previousKeyState[scancode] == 0 {
				InputChannel <- InputEvent{KeyCode: keyCode}
			}
		}

		copy(previousKeyState, currentKeyState)
		sdl.Delay(50)
	}
}

func listenForControllerEvents() {
	controller := openController()
	defer controller.Close()

	controllerMappings := map[sdl.GameControllerButton]string{
		sdl.CONTROLLER_BUTTON_DPAD_DOWN: "DOWN",
		sdl.CONTROLLER_BUTTON_DPAD_UP:   "UP",
		sdl.CONTROLLER_BUTTON_A:         "B",
		sdl.CONTROLLER_BUTTON_B:         "A",
		sdl.CONTROLLER_BUTTON_X:         "X",
		sdl.CONTROLLER_BUTTON_Y:         "Y",
		sdl.CONTROLLER_BUTTON_START:     "START",
		sdl.CONTROLLER_BUTTON_BACK:      "SELECT",
		sdl.CONTROLLER_BUTTON_GUIDE:     "MENU",
	}

	// State tracking for debounce
	previousButtonState := make(map[sdl.GameControllerButton]bool)

	for {
		sdl.PumpEvents()
		for button, keyCode := range controllerMappings {
			currentState := controller.Button(button) == sdl.PRESSED
			if currentState && !previousButtonState[button] {
				InputChannel <- InputEvent{KeyCode: keyCode}
			}
			previousButtonState[button] = currentState
		}

		sdl.Delay(50)
	}
}

func openController() *sdl.GameController {
	for i := 0; i < sdl.NumJoysticks(); i++ {
		if sdl.IsGameController(i) {
			controller := sdl.GameControllerOpen(i)
			if controller != nil {
				return controller
			}
		}
	}
	return nil
}
