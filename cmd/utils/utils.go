package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/ebarthur/jotl/cmd/config"
	"github.com/ebarthur/jotl/cmd/db"
	"github.com/ebarthur/jotl/cmd/flags"
	"github.com/gofrs/flock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/pflag"
)

const envTemplate = `# Jotl is a modern CLI tool designed to streamline log management 
# for developers like you.
# Please consider starring the repo if you find it useful: https://github.com/ebarthur/jotl

DB_CONNECTION_STRING=%s

# Application Configuration (This is read-only)
APP_NAME=%s
`

type ConfigPaths struct {
	ConfigDir  string
	ConfigFile string
	DBDir      string
}

var (
	postgresDriver = []string{"github.com/lib/pq"}
	sqliteDriver   = []string{"github.com/mattn/go-sqlite3"}
)

const (
	lockFileName = ".jotl.lock"
	ProgramName  = "jotl"
)

// GetConfigPaths returns the necessary paths for configuration
func GetConfigPaths(currentDir string) ConfigPaths {
	return ConfigPaths{
		ConfigDir:  filepath.Join(currentDir, "jotl"),
		ConfigFile: filepath.Join(currentDir, "jotl", "config.yaml"),
		DBDir:      filepath.Join(currentDir, "jotl", "db"),
	}
}

// ValidateModuleName checks if the provided module name is valid.
// A valid module name can contain alphanumeric characters, underscores,
// hyphens, and can be separated by dots or slashes.
// It returns true if it's a valid module name.
func ValidateModuleName(moduleName string) bool {
	matched, _ := regexp.Match("^[a-zA-Z0-9_-]+(?:[\\/.][a-zA-Z0-9_-]+)*$", []byte(moduleName))
	return matched
}

func CheckGitConfig(key string) (bool, error) {
	cmd := exec.Command("git", "config", "--get", key)
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// The command failed to run.
			if exitError.ExitCode() == 1 {
				// The 'git config --get' command returns 1 if the key was not found.
				return false, nil
			}
		}
		// Some other error occurred
		return false, err
	}
	// The command ran successfully, so the key is set.
	return true, nil
}

// IsGitDirectory checks if the given directory is a Git repository.
// It returns true if the directory contains a .git directory or a .git file
// (which can be the case for git submodules or alternative git configurations).
// It returns false otherwise.
func IsGitDirectory(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		return true
	}

	// Check for .git file (in case of git submodules or alternative git configs)
	gitFile := filepath.Join(dir, ".git")
	if info, err := os.Stat(gitFile); err == nil && !info.IsDir() {
		return true
	}

	return false
}

// NonInteractiveCommand creates the command string from a flagSet
// to be used for getting the equivalent non-interactive shell command
func NonInteractiveCommand(use string, flagSet *pflag.FlagSet) string {
	nonInteractiveCommand := fmt.Sprintf("%s %s", ProgramName, use)

	visitFn := func(flag *pflag.Flag) {
		if flag.Name != "help" {
			if flag.Name == "feature" {
				featureFlagsString := ""
				// Creates string representation for the feature flags to be
				// concatenated with the nonInteractiveCommand
				for _, k := range strings.Split(flag.Value.String(), ",") {
					if k != "" {
						featureFlagsString += fmt.Sprintf(" --feature %s", k)
					}
				}
				nonInteractiveCommand += featureFlagsString
			} else if flag.Value.Type() == "bool" {
				if flag.Value.String() == "true" {
					nonInteractiveCommand = fmt.Sprintf("%s --%s", nonInteractiveCommand, flag.Name)
				}
			} else {
				nonInteractiveCommand = fmt.Sprintf("%s --%s %s", nonInteractiveCommand, flag.Name, flag.Value.String())
			}
		}
	}

	flagSet.SortFlags = false
	flagSet.VisitAll(visitFn)

	return nonInteractiveCommand
}

// ExecuteCmd provides a shorthand way to run a shell command
func ExecuteCmd(name string, args []string, dir string) error {
	command := exec.Command(name, args...)
	command.Dir = dir
	var out bytes.Buffer
	command.Stdout = &out
	if err := command.Run(); err != nil {
		return err
	}
	return nil
}

// EnsureGitignore checks if a .gitignore file exists in the given directory,
// creates it if it does not exist, and ensures that the /jotl directory is ignored.
func EnsureGitignore(dir string) error {
	gitignorePath := filepath.Join(dir, ".gitignore")

	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		file, err := os.Create(gitignorePath)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		return err
	}

	if !strings.Contains(string(content), "/jotl") {
		file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := file.WriteString("\n/jotl\n"); err != nil {
			return err
		}
	}

	return nil
}

// InstallDatabaseDrivers installs the specified database driver package using the `go get` command.
// Supported database drivers are "postgres" and "sqlite".
func InstallDatabaseDrivers(dbDriver string) error {
	var driverPackage string

	switch dbDriver {
	case "postgres":
		driverPackage = sqliteDriver[0]
	case "sqlite":
		driverPackage = postgresDriver[0]
	default:
		return fmt.Errorf("unsupported database driver: %s", dbDriver)
	}

	if err := ExecuteCmd("go", []string{"get", "-u", driverPackage}, ""); err != nil {
		return fmt.Errorf("failed to install database driver %s: %w", driverPackage, err)
	}

	return nil
}

// CreateDatabase sets up the database for the jotl project based on the specified driver.
// For SQLite, it creates a database file in the jotl directory.
// For PostgreSQL, it assumes the database is managed externally and does not create any files (well except for a docker-compose.yaml).
func CreateDatabase(currentDir string, dbDriver string) error {

	switch dbDriver {
	case "sqlite":
		configPaths := GetConfigPaths(currentDir)
		sqliteDB := configPaths.DBDir + "/jotl.db"

		_, err := db.NewLogsDB(sqliteDB)
		if err != nil {
			return fmt.Errorf("failed to create sqlite database: %w", err)
		}

	case "postgres":
		// For PostgreSQL, we don't need to create any db file
		// The connection string will be used when connecting to the database
		// Touch docker-compose.yml for the postgres
		config := config.DefaultPostgresConfig()
		if err := config.CreateDockerCompose(currentDir, config); err != nil {
			return fmt.Errorf("failed to create docker-compose configuration: %w", err)
		}
	default:
		return fmt.Errorf("unsupported database driver: %s", dbDriver)
	}

	return nil
}

// InitializeJotlDirectory creates the necessary directory structure and files for jotl.
// It creates the following structure:
// root directory/
//
//			└── jotl/
//	    		├── db/
//	    		│   ├── .env
//	    		│   └── jotl.db
//	    		└── jotl.config.yaml
//
// This structure is exclusive of postgres driver
func InitializeJotlDirectory(currentDir, dbDriver string) error {
	paths := GetConfigPaths(currentDir)

	if err := os.MkdirAll(paths.ConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create jotl directory: %w", err)
	}

	if dbDriver != "postgres" {
		if err := os.MkdirAll(paths.DBDir, 0755); err != nil {
			return fmt.Errorf("failed to create db directory: %w", err)
		}
	}

	return nil
}

// IsJotlInitialized checks whether a jotl project has been initialized already in the directory tree.
// It returns true if the jotl directory and its necessary files exist.
func IsJotlInitialized(startDir string) bool {
	// Directories to skip for performance
	skipDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
	}

	for {
		entries, err := os.ReadDir(startDir)
		if err != nil {
			break
		}

		for _, entry := range entries {
			if entry.IsDir() {
				if skipDirs[entry.Name()] {
					continue
				}

				if entry.Name() == "jotl" {
					paths := GetConfigPaths(startDir)

					if _, err := os.Stat(paths.ConfigDir); err == nil {

						if _, err := os.Stat(paths.ConfigFile); err == nil {
							return true
						} else {
						}
					} else {
					}
				}
			}
		}

		// Move up one directory
		parentDir := filepath.Dir(startDir)
		if parentDir == startDir { // Reached root directory
			break
		}
		startDir = parentDir
	}

	return false
}

// CreateEnvFile creates a .env file with the configuration
func CreateEnvFile(currentDir string, cfg *config.JotlConfig) error {
	envPath := GetConfigPaths(currentDir).ConfigDir + "/.env"

	envContent := fmt.Sprintf(
		envTemplate,
		cfg.Database.Path,
		cfg.Project.Name,
	)

	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to create .env file: %w", err)
	}

	return nil
}

// DeleteJotlOnFail deletes the Jotl configuration directory if an error occurs.
func DeleteJotlOnFail(currentDir string) error {
	paths := GetConfigPaths(currentDir)
	if err := os.RemoveAll(paths.ConfigDir); err != nil {
		return fmt.Errorf("failed to delete jotl directory: %w", err)
	}
	return nil
}

// GetDevEnvironment scans the project directory to detect the language server
func GetDevEnvironment() string {
	var devEnv string

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			switch filepath.Base(path) {
			case "go.mod":
				devEnv = "go"
				return filepath.SkipDir
			case "package.json":
				devEnv = "node"
				return filepath.SkipDir
			case "requirements.txt":
				devEnv = "python"
				return filepath.SkipDir
			}
		}
		return nil
	})

	if devEnv == "" {
		devEnv = "development"
	}

	return devEnv
}

// AcquireLock attempts to create and lock a session lock file
// This prevents multiple instances of `jotl dev`
func AcquireLock() (*flock.Flock, error) {
	// Define the lock file path in the temporary directory
	lockFilePath := filepath.Join(os.TempDir(), lockFileName)

	// Create a flock instance
	fileLock := flock.New(lockFilePath)

	// Attempt to acquire an exclusive, non-blocking lock
	locked, err := fileLock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return nil, errors.New("another session of 'jotl dev' is already running")
	}

	return fileLock, nil
}

// ReleaseLock releases and deletes the session lock file
// We employ this when after we run a `jotl dev` instance
func ReleaseLock(fl *flock.Flock) error {
	if err := fl.Unlock(); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if err := os.Remove(fl.Path()); err != nil {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	return nil
}

// FindAvailablPort finds the perfect port for the web studio
func FindAvailablePort(startPort int) string {
	for port := startPort; port <= flags.MaxPort; port++ {
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			return strconv.Itoa(port)
		}
	}
	log.Fatal("No available ports found")
	return ""
}

// ParsePort converts a string port to an integer.
func ParsePort(port string) int {
	parsedPort, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Invalid port: %s", port)
	}
	return parsedPort
}

// OpenBrowser attempts to open the default web browser with the given URL.
// We use this to automatically open the web studio
func OpenBrowser(url string) {
	var err error

	switch os := runtime.GOOS; os {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}

// User config is accessed in web studio by serving it on an endpoint
// that the static app can access
func ServeConfig(port string) error {
	configPort := ParsePort(port) + 1
	addr := fmt.Sprintf(":%d", configPort)

	http.HandleFunc("/cfg", func(w http.ResponseWriter, r *http.Request) {
		currentDir, _ := os.Getwd()
		cfg, err := config.LoadConfig(currentDir)
		if err != nil {
			http.Error(w, "Failed to load configuration", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cfg); err != nil {
			http.Error(w, "Failed to encode configuration", http.StatusInternalServerError)
		}
	})

	if err := http.ListenAndServe(addr, nil); err != nil {
		return fmt.Errorf("config server failed: %w", err)
	}
	return nil
}
