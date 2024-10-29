package tracer

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

func NewStdoutLogger() logr.Logger {
	return funcr.New(func(prefix, args string) {
		if prefix != "" {
			fmt.Printf("logr: %s: %s\n", prefix, args)
		} else {
			fmt.Println(args)
		}
	}, funcr.Options{})
}
