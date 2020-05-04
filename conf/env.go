package conf

import (
	"os"
	"strings"
)

// GetEnv get runtime env
func GetEnv() string {
	if len(os.Args) > 0 && strings.Contains(os.Args[0], ".test") {
		return "TEST"
	}

	return "PROD"
}
