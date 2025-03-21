package output

import (
	"fmt"

	"github.com/anibaldeboni/screech/config"
)

func Printf(format string, a ...any) {
	if config.Debug {
		fmt.Printf(format, a...)
	}
}
