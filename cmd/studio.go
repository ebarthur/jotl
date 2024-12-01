package cmd

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/ebarthur/jotl/cmd/config"
	"github.com/ebarthur/jotl/cmd/flags"
	"github.com/ebarthur/jotl/cmd/utils"
	"github.com/spf13/cobra"
)

//go:embed gui/build/dist/*
var static embed.FS

var port flags.Port

const lngMessage = (`The studio command launches a web-based dashboard for Jotl.

It provides a modern, user-friendly interface to:
- View and filter logs across different environments
- Search and analyze logs by timestamp, status codes, and messages
- Monitor real-time log updates through the dashboard
- Export and share log data

The dashboard automatically starts on port 8080 and will increment
until it finds an available port if 8080 is in use.

Note: The studio dashboard requires the project to be initialized with 'jotl init'
and have a valid database connection configured in the jotl directory.`)

var studioCommand = &cobra.Command{
	Use:   "studio",
	Short: "Start a local web server to view logs",
	Long: func() string {
		out, _ := glamour.Render(lngMessage, "dark")
		return out
	}(),

	Run: func(cmd *cobra.Command, args []string) {
		portFlag, _ := cmd.Flags().GetString("port")
		currentDir, _ := os.Getwd()

		if !utils.IsJotlInitialized(currentDir) {
			fmt.Printf("%s\n", logoStyle.Render(logo))
			fmt.Println(endingMsgStyle.Render("\nYou need to initialize a Jotl project first. Run `jotl init` on your project root directory!"))

			return
		}

		if portFlag == "" {
			cfg, err := config.LoadConfig(currentDir)
			if err != nil {
				fmt.Printf("%s\n", logoStyle.Render(logo))
				fmt.Println(endingMsgStyle.Render("\nOops, there is something wrong with your jotl.config.yaml. Verify config and try again."))
				return
			}

			if err := port.Set(cfg.Dashboard.Port); err != nil {
				fmt.Println(endingMsgStyle.Render("Invalid port in configuration. Using default port: 8080"))
				port.Set("8080")
			}
		} else {
			if err := port.Set(portFlag); err != nil {
				fmt.Println(endingMsgStyle.Render("Invalid port provided. Using default port: 8080"))
				port.Set("8080")
			}
		}

		finalPort := utils.FindAvailablePort(utils.ParsePort(port.String()))

		startServer(finalPort)
	},
}

func init() {
	studioCommand.Flags().StringVarP((*string)(&port), "port", "p", "8080", "Port to run the studio dashboard (default: 8080)")
	rootCmd.AddCommand(studioCommand)
}

func startServer(port string) {
	go utils.ServeConfig(port)

	stripped, err := fs.Sub(static, "gui/build/dist")
	if err != nil {
		log.Fatal(err)
	}

	staticHandler := http.FileServer(http.FS(stripped))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f, err := stripped.Open(r.URL.Path)
		if os.IsNotExist(err) {
			r.URL.Path = "/"
		} else if err == nil {
			f.Close()
		}
		staticHandler.ServeHTTP(w, r)
	})

	apiServer := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}

	done := make(chan bool, 1)
	go gracefulShutdown(apiServer, done)

	link := "http://localhost:" + port
	fmt.Printf("%s\n", logoStyle.Render(logo))
	fmt.Println(tipMsgStyle.Render(fmt.Sprintf("Server starting on: %s", clickableLink(link)))) // change stule

	go func() {
		time.Sleep(2 * time.Second)
		utils.OpenBrowser(link)
	}()

	if err := apiServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %v", err)
	}

	<-done
	fmt.Println(endingMsgStyle.Render("Server stopped"))
}

// This makes the link clickable in the terminal (Though we automatically open the browser for the user)
func clickableLink(url string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, url)
}

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	fmt.Println(endingMsgStyle.Render("shutting down gracefully, press Ctrl+C again to force"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	fmt.Println(endingMsgStyle.Render("Server exiting"))

	done <- true
}
