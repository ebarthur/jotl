package program

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ebarthur/jotl/cmd/flags"
	"github.com/ebarthur/jotl/cmd/utils"
)

type Project struct {
	ProjectName  string
	Exit         bool
	AbsolutePath string
	LogLevel     flags.LogLevel
	DBDriver     flags.Database
	GitOptions   flags.Git
	OSCheck      map[string]bool
}

// var (
// 	postgresDriver = []string{"github.com/lib/pq"}
// 	sqliteDriver   = []string{"github.com/mattn/go-sqlite3"}
// )

// ExitCLI checks if the Project has been exited, and closes
// out of the CLI if it has
func (p *Project) ExitCLI(tprogram *tea.Program) {
	if p.Exit {
		// logo render here
		if err := tprogram.ReleaseTerminal(); err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	}
}

// CreateJotlProject sets up a new Jotl project in the current directory.
// It checks if the user's git email is configured, initializes a git repository if needed,
// creates necessary directories and files, and installs database drivers.
func (p *Project) CreateJotlProject() error {

	//check if user.email is set.
	emailSet, err := utils.CheckGitConfig("user.email")
	if err != nil {
		return err
	}

	// if user.email is not set, prompt user to set before continuing
	if !emailSet && p.GitOptions.IsGitEnabled() {
		fmt.Println("user.email is not set in git config.")
		fmt.Println("Please set up git config before trying again.")
		panic("\nGIT CONFIG ISSUE: user.email is not set in git config.\n")
	}

	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %v", err)
		return err
	}

	if p.GitOptions.IsGitEnabled() {
		isGitRepo := utils.IsGitDirectory(currentDir)
		// if git repo return, if not, check for --git flag and initialize
		if !isGitRepo {
			log.Printf("Error initializing git repo in directory %s: %v", currentDir, err)
			err = utils.ExecuteCmd("git", []string{"init"}, currentDir)

			if err != nil {
				log.Printf("Error initializing git repo: %v", err)
				return err
			}
		}

		// create .gitignore if it does not exist, check if /jotl is ignored
		if err := utils.EnsureGitignore(currentDir); err != nil {
			log.Printf("Error ensuring .gitignore file in directory %s: %v", currentDir, err)
		}

		err := utils.InitializeJotlDirectory(currentDir)
		if err != nil {
			return fmt.Errorf("failed to initialize jotl directory: %w", err)
		}
		// install db drivers

		// set config schema

	}
	return nil

}
