/*
Copyright © 2024 Ebenezer Arthur arthurebenezer@aol.com
*/
package cmd

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/spf13/cobra"
)

// JotlVersion is the version of the cli. This may be later overwritten by goreleaser in the CI run with the version of the release on github
var JotlVersion string

// getJotlVersion retrieves the version information of the Jotl application.
// It first checks if the JotlVersion variable is set and returns it if available.
// If not, it attempts to read build information using debug.ReadBuildInfo().
// If the build information is available and the main version is not "(devel)",
// it returns the main version. Otherwise, it looks for VCS revision and time
// settings in the build information and formats them accordingly.
// If no version information is available, it returns a default message indicating
// that no version information is available for the build.
func getJotlVersion() string {
	noVersionAvailable := "No version information available for this build. Run 'jotl help version' for more information."

	if len(JotlVersion) != 0 {
		return JotlVersion
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return noVersionAvailable
	}

	// If no main version is available, Go defaults it to (devel)
	if buildInfo.Main.Version != "(devel)" {
		return buildInfo.Main.Version
	}

	var vcsRevision string
	var vcsTime time.Time
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.revision":
			vcsRevision = setting.Value
		case "vcs.time":
			var err error
			vcsTime, err = time.Parse(time.RFC3339, setting.Value)
			if err != nil {
				return fmt.Sprintf("Invalid time format: %v", err)
			}
			return fmt.Sprintf("Revision: %s, Date: %s", vcsRevision, vcsTime)
		}
	}

	if vcsRevision != "" {
		return fmt.Sprintf("%s, (%s)", vcsRevision, vcsTime)
	}

	return noVersionAvailable
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display application version information.",
	Long: `
The version command provides information about the application's version.

Jotl requires version information to be embedded at compile time.
For detailed version information, Jotl needs to be built as specified in the README installation instructions.
If Jotl is built within a version control repository and other version info isn't available,
the revision hash will be used instead.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		version := getJotlVersion()
		fmt.Printf("Jotl CLI version: %v\n", version)
	},
}
