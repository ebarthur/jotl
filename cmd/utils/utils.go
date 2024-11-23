package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/pflag"
)

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

const ProgramName = "jotl"

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

	// Check if .gitignore file exists, create if it does not
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		file, err := os.Create(gitignorePath)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	// Read the .gitignore file
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		return err
	}

	// Check if /jotl is already ignored
	if !strings.Contains(string(content), "/jotl") {
		// Append /jotl to .gitignore
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

// InitializeJotlDirectory creates the necessary directory structure and files for jotl.
// It creates the following structure:
// currentDir/
//
//			└── jotl/
//	    		├── db/
//	    		│   ├── .env
//	    		│   └── jotl.db
//	    		└── jotl.config.yaml
func InitializeJotlDirectory(currentDir string) error {
	jotlDir := filepath.Join(currentDir, "jotl")
	dbDir := filepath.Join(jotlDir, "db")

	// Create jotl and db directories
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// List of files to create with their respective paths
	filesToCreate := []string{
		filepath.Join(jotlDir, "jotl.config.yaml"),
		filepath.Join(dbDir, ".env"),
		filepath.Join(dbDir, "jotl.db"),
	}

	// Create each file in the list
	for _, filePath := range filesToCreate {
		if err := createFile(filePath); err != nil {
			return err
		}
	}

	return nil
}

// createFile creates an empty file at the specified path.
func createFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()
	return nil
}
