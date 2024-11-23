package program

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ebarthur/jotl/cmd/config"
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
func (p *Project) CreateJotlProject(currentDir string, cfg config.JotlConfig) error {

	//check if user.email is set.
	emailSet, _ := utils.CheckGitConfig("user.email")

	// if user.email is not set, prompt user to set before continuing
	if !emailSet && p.GitOptions.IsGitEnabled() {
		fmt.Println("user.email is not set in git config.")
		fmt.Println("Please set up git config before trying again.")
		panic("\nGIT CONFIG ISSUE: user.email is not set in git config.\n")
	}

	// if user chose "Yes" to Git options
	if p.GitOptions.IsGitEnabled() {
		isGitRepo := utils.IsGitDirectory(currentDir)
		// if directory is not git repo, initialize
		if !isGitRepo {
			err := utils.ExecuteCmd("git", []string{"init"}, currentDir)
			if err != nil {
				log.Printf("Error initializing git repo: %v", err)
				return err
			}
		}

		// create .gitignore if it does not exist, check if /jotl is ignored
		if err := utils.EnsureGitignore(currentDir); err != nil {
			log.Printf("Error ensuring .gitignore file in directory %s: %v", currentDir, err)
		}
	}

	// Initialize Jotl directory structure
	if err := utils.InitializeJotlDirectory(currentDir, p.DBDriver.String()); err != nil {
		return err
	}

	// Save configuration to yaml file
	if err := cfg.SaveConfig(utils.GetConfigPaths(currentDir).ConfigFile); err != nil {
		return err
	}

	// Install the necessary DB Drivers
	// `go get -` is used
	// what if user doesn't have go on their machine? []: TODO: Support npm for installation
	if err := utils.InstallDatabaseDrivers(string(p.DBDriver)); err != nil {
		return err
	}

	// Initialize the `SQLite` db in directory
	// For `postgres`, do nothing (the connection url is used to connect to external dbs)
	if err := utils.CreateDatabase(currentDir, string(p.DBDriver)); err != nil {
		return err
	}

	// Create env file for the db and other configs
	if err := utils.CreateEnvFile(currentDir, &cfg); err != nil {
		return err
	}

	return nil
}
