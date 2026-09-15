package main

import (
	"fmt"
	"os"

	"github.com/SamandarMadaliev/ex-rate/cmd/app"
)

// @title       Ex-Rate API
// @version     1.0
// @description Currency exchange rate service that fetches and stores rates for currency pairs.
// @BasePath    /api/v1
func main() {
	if err := app.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
