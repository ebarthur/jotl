/*
Copyright © 2024 Ebenezer Arthur arthurebenezer@aol.com
*/

package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ebarthur/jotl/cmd/flags"
	"github.com/ebarthur/jotl/cmd/program"
	"github.com/ebarthur/jotl/cmd/steps"
	multiInput "github.com/ebarthur/jotl/cmd/ui/multi_input"
	"github.com/ebarthur/jotl/cmd/ui/spinner"
	textinput "github.com/ebarthur/jotl/cmd/ui/text_input"
	"github.com/ebarthur/jotl/cmd/utils"
	"github.com/spf13/cobra"
)

const logo = `
    ___      ___    ___      ___   
   /\  \    /\  \  /\  \    /\__\  
  _\:\  \  /::\  \ \:\  \  /:/  /  
 /\/::\__\/:/\:\__\/::\__\/:/__/   
 \::/\/__/\:\/:/  /:/\/__/\:\  \   
  \/__/    \::/  /\/__/    \:\__\  
            \/__/           \/__/  

`

var (
	logoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#01FAC6")).Bold(true)
	tipMsgStyle = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("190")).Italic(true)
	// endingMsgStyle = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("170")).Bold(true)
)

func init() {
	var flagDBDriver flags.Database
	var flagLogLevel flags.LogLevel
	var flagGit bool
	rootCmd.AddCommand(initCommand)

	initCommand.Flags().StringP("name", "n", "", "Name of Jotl project")
	initCommand.Flags().VarP(&flagDBDriver, "driver", "d", fmt.Sprintf("Database drivers to use. Allowed values: %s", strings.Join(flags.AllowedDBDrivers, ", ")))
	initCommand.Flags().VarP(&flagLogLevel, "log", "l", fmt.Sprintf("Log levels available. Allowed configs: %s", strings.Join(flags.AllowedLogLevels, ", ")))
	initCommand.Flags().BoolVarP(&flagGit, "git", "g", false, "Initialize Git repository (True/False)")
}

type Options struct {
	ProjectName *textinput.Output
	DBDriver    *multiInput.Selection
	LogLevel    *multiInput.Selection
	Git         *multiInput.Selection
}

var initCommand = &cobra.Command{
	Use:   "init",
	Short: "Initializes a Jotl project in your root directory",
	Long: `The init command sets up a new Jotl project in your root directory.
It creates the necessary configuration files and directory structure
to get you started with Jotl. Make sure to run this command in the
directory where you want to initialize your project.`,

	Run: func(cmd *cobra.Command, args []string) {
		var tprogram *tea.Program
		var err error

		isInteractive := false
		flagName := cmd.Flag("name").Value.String()

		if flagName != "" && !utils.ValidateModuleName(flagName) {
			err = fmt.Errorf("'%s' is not a valid module name. Please choose a different name", flagName)
			cobra.CheckErr(textinput.CreateErrorInputModel(err).Err())
		}

		// VarP already validates the contents of the framework flag.
		// If this flag is filled, it is always valid
		flagLogLevel := flags.LogLevel(cmd.Flag("log").Value.String())
		flagDBDriver := flags.Database(cmd.Flag("driver").Value.String())
		flagGit := cmd.Flag("git").Value.String() == "true"

		options := Options{
			ProjectName: &textinput.Output{},
			DBDriver:    &multiInput.Selection{},
			LogLevel:    &multiInput.Selection{},
			Git:         &multiInput.Selection{},
		}

		project := &program.Project{
			ProjectName: flagName,
			LogLevel:    flagLogLevel,
			DBDriver:    flagDBDriver,
			GitOptions:  flags.Git(flagGit),
		}

		steps := steps.InitSteps(flagDBDriver, flagLogLevel)
		fmt.Printf("%s\n", logoStyle.Render(logo))

		if project.ProjectName == "" {
			isInteractive = true
			tprogram := tea.NewProgram(textinput.InitialTextInputModel(options.ProjectName, "Name your Jotl project.", project)) // personalize header using git config
			if _, err := tprogram.Run(); err != nil {
				log.Printf("Name of project contains an error: %v", err)
				cobra.CheckErr(textinput.CreateErrorInputModel(err).Err())
			}

			if options.ProjectName.Output != "" && !utils.ValidateModuleName(options.ProjectName.Output) {
				err = fmt.Errorf("'%s' is not a valid module name. Please choose a different name", options.ProjectName.Output)
				cobra.CheckErr(textinput.CreateErrorInputModel(err).Err())
			}

			project.ExitCLI(tprogram)

			// TODO: Set config name
			project.ProjectName = options.ProjectName.Output
			err := cmd.Flag("name").Value.Set(project.ProjectName)
			if err != nil {
				log.Fatal("failed to set the name flag value", err)
			}
		}

		if project.DBDriver == "" {
			isInteractive = true
			step := steps.Steps["driver"]
			tprogram = tea.NewProgram(multiInput.InitialModelMulti(step.Options, options.DBDriver, step.Headers, project))
			if _, err := tprogram.Run(); err != nil {
				cobra.CheckErr(textinput.CreateErrorInputModel(err).Err())
			}
			project.ExitCLI(tprogram)

			// this type casting is always safe since the user interface can
			// only pass strings that can be cast to a flags.Database instance
			project.DBDriver = flags.Database(strings.ToLower(options.DBDriver.Choice))
			err := cmd.Flag("driver").Value.Set(project.DBDriver.String())
			if err != nil {
				log.Fatal("failed to set the driver flag value", err)
			}
		}

		if project.LogLevel == "" {
			step := steps.Steps["log_level"]
			tprogram = tea.NewProgram(multiInput.InitialModelMulti(step.Options, options.LogLevel, step.Headers, project))
			if _, err := tprogram.Run(); err != nil {
				cobra.CheckErr(textinput.CreateErrorInputModel(err).Err())
			}
			project.ExitCLI(tprogram)

			// this type casting is always safe since the user interface can
			// only pass strings that can be cast to a flags.LogLevel instance
			project.LogLevel = flags.LogLevel(strings.ToLower(options.LogLevel.Choice))
			err := cmd.Flag("log").Value.Set(project.LogLevel.String())
			if err != nil {
				log.Fatal("failed to set the driver flag value", err)
			}
		}

		if !project.GitOptions {
			isInteractive = true
			step := steps.Steps["git"]
			tprogram = tea.NewProgram(multiInput.InitialModelMulti(step.Options, options.Git, step.Headers, project))
			if _, err := tprogram.Run(); err != nil {
				cobra.CheckErr(textinput.CreateErrorInputModel(err).Err())
			}
			project.ExitCLI(tprogram)

			gitEnabled := options.Git.Choice == "Yes"
			project.GitOptions = flags.Git(gitEnabled)
			err := cmd.Flag("git").Value.Set(project.GitOptions.String())
			if err != nil {
				log.Fatal("failed to set the git flag value", err)
			}
		}

		currentWorkingDir, err := os.Getwd()
		if err != nil {
			log.Printf("could not get current working directory: %v", err)
			cobra.CheckErr(textinput.CreateErrorInputModel(err).Err())
		}
		project.AbsolutePath = currentWorkingDir

		spinner := tea.NewProgram(spinner.InitialModelNew())

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := spinner.Run(); err != nil {
				cobra.CheckErr(err)
			}
		}()

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("The program encountered an unexpected issue and had to exit. The error was:", r)
				fmt.Println("If you continue to experience this issue, please post a message on our GitHub page or dm on X @StatmanAartt")
				if releaseErr := spinner.ReleaseTerminal(); releaseErr != nil {
					log.Printf("Problem releasing terminal: %v", releaseErr)
				}
			}
		}()

		// []: Work on this, refactor (this is just a test run for CreateJotlProject)
		// Also handle when a user tries to initialize a project
		err = project.CreateJotlProject()
		if err != nil {
			if releaseErr := spinner.ReleaseTerminal(); releaseErr != nil {
				log.Printf("Problem releasing terminal: %v", releaseErr)
			}
			log.Printf("Problem creating files for project. %v", err)
			cobra.CheckErr(textinput.CreateErrorInputModel(err).Err())
		}

		if isInteractive {
			nonInteractiveCommand := utils.NonInteractiveCommand(cmd.Use, cmd.Flags())
			fmt.Println(tipMsgStyle.Render("Tip: Repeat the equivalent Jotl with the following non-interactive command:"))
			fmt.Println(tipMsgStyle.Italic(false).Render(fmt.Sprintf("• %s\n", nonInteractiveCommand)))
		}
		err = spinner.ReleaseTerminal()
		if err != nil {
			log.Printf("Could not release terminal: %v", err)
			cobra.CheckErr(err)
		}
	},
}
