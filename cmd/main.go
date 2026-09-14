package main

import (
	"fmt"
	"os"

	"github.com/SamandarMadaliev/ex-rate/cmd/app"
)

func main() {
	if err := app.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
