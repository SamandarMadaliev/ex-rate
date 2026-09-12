package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/SamandarMadaliev/ex-rate/internal/app"
	"github.com/SamandarMadaliev/ex-rate/pkg/config"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "http-server",
	Short: "Start HTTP Server",
	Run: func(cmd *cobra.Command, args []string) {
		_ = godotenv.Load()

		cfg, err := config.NewConfig()
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}

		app, err := app.NewApp(cfg)
		if err != nil {
			log.Fatalf("failed to create app: %v", err)
		}

		go func() {
			if err := app.Run(); err != nil {
				if err == http.ErrServerClosed {
					return
				}

				log.Println("app run", err)
			}
		}()

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs

		// app stops
		log.Println("application server stopping...")
		_ = app.Stop()
		log.Println("application backend server stopped gracefully")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
