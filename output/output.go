package output

import (
	"fmt"

	"github.com/anibaldeboni/screech/config"
)

func Printf(format string, a ...any) (n int, err error) {
	if config.Debug {
		return fmt.Printf(format, a...)
	}
	return 0, nil
}
