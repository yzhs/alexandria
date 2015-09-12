package util

import (
	"fmt"
	"os"
)

func LogError(err interface{}) {
	fmt.Fprintln(os.Stderr, err)
}

func TryLogError(err interface{}) {
	if err != nil {
		LogError(err)
	}
}
