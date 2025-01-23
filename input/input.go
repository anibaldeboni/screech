package input

import (
	"runtime"

	"github.com/veandco/go-sdl2/sdl"
)

type UserInputEvent struct {
	KeyCode string
}

var UserInputChannel = make(chan UserInputEvent)

func StartListening() {
	go listenForKeyboardEvents()
	if runtime.GOOS == "linux" {
		go listenForControllerEvents()
	}
}

func listenForKeyboardEvents() {
	previousKeyState := make([]uint8, sdl.NUM_SCANCODES)

	keyMappings := map[sdl.Scancode]string{
		sdl.SCANCODE_DOWN:   "DOWN",
		sdl.SCANCODE_UP:     "UP",
		sdl.SCANCODE_A:      "A",
		sdl.SCANCODE_B:      "B",
		sdl.SCANCODE_X:      "X",
		sdl.SCANCODE_RETURN: "START",
		sdl.SCANCODE_ESCAPE: "SELECT",
	}

	for {
		currentKeyState := sdl.GetKeyboardState()

		for scancode, keyCode := range keyMappings {
			isClickingMappedKey := currentKeyState[scancode] != 0
			isNotClickingAgain := previousKeyState[scancode] == 0
			shouldSendEvent := isNotClickingAgain || isAllowedRepeatableKey(scancode, sdl.SCANCODE_DOWN, sdl.SCANCODE_UP)

			if isClickingMappedKey && shouldSendEvent {
				UserInputChannel <- UserInputEvent{KeyCode: keyCode}
			}
		}
		copy(previousKeyState, currentKeyState)
		sdl.Delay(150)
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
		sdl.CONTROLLER_BUTTON_X:         "Y",
		sdl.CONTROLLER_BUTTON_Y:         "X",
		sdl.CONTROLLER_BUTTON_START:     "START",
		sdl.CONTROLLER_BUTTON_BACK:      "SELECT",
		sdl.CONTROLLER_BUTTON_GUIDE:     "MENU",
	}

	// State tracking for debounce
	previousButtonState := make(map[sdl.GameControllerButton]bool)

	for {
		sdl.PumpEvents()
		for button, keyCode := range controllerMappings {
			isClickingMappedKey := controller.Button(button) == sdl.PRESSED
			isNotClickingAgain := !previousButtonState[button]
			shouldSendEvent := isNotClickingAgain || isAllowedRepeatableKey(sdl.Scancode(button), sdl.CONTROLLER_BUTTON_DPAD_DOWN, sdl.CONTROLLER_BUTTON_DPAD_UP)

			if isClickingMappedKey && shouldSendEvent {
				UserInputChannel <- UserInputEvent{KeyCode: keyCode}
			}
			previousButtonState[button] = isClickingMappedKey
		}

		sdl.Delay(150)
	}
}

func isAllowedRepeatableKey(current sdl.Scancode, allowed ...sdl.Scancode) bool {
	for _, a := range allowed {
		if current == a {
			return true
		}
	}
	return false
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
