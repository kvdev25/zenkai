package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"cli/zkcli/internal/theme"
)

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage themes",
}

var applyCmd = &cobra.Command{
	Use:   "apply [name]",
	Short: "Apply a theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		configBase, themesDir, templatesDir, err := resolvePaths()
		if err != nil {
			return err
		}

		return theme.ApplyThemeByName(
			configBase,
			themesDir,
			templatesDir,
			args[0],
		)
	},
}

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Apply next theme",
	RunE: func(cmd *cobra.Command, args []string) error {

		configBase, themesDir, templatesDir, err := resolvePaths()
		if err != nil {
			return err
		}

		next, err := theme.GetNextTheme(configBase, themesDir)
		if err != nil {
			return err
		}

		return theme.ApplyThemeByName(
			configBase,
			themesDir,
			templatesDir,
			next,
		)
	},
}

var prevCmd = &cobra.Command{
	Use:   "prev",
	Short: "Apply previous theme",
	RunE: func(cmd *cobra.Command, args []string) error {

		configBase, themesDir, templatesDir, err := resolvePaths()
		if err != nil {
			return err
		}

		prev, err := theme.GetPrevTheme(configBase, themesDir)
		if err != nil {
			return err
		}

		return theme.ApplyThemeByName(
			configBase,
			themesDir,
			templatesDir,
			prev,
		)
	},
}

var randomCmd = &cobra.Command{
	Use:   "random",
	Short: "Apply random theme",
	RunE: func(cmd *cobra.Command, args []string) error {

		configBase, themesDir, templatesDir, err := resolvePaths()
		if err != nil {
			return err
		}

		r, err := theme.GetRandomTheme(themesDir)
		if err != nil {
			return err
		}

		return theme.ApplyThemeByName(
			configBase,
			themesDir,
			templatesDir,
			r,
		)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available themes",
	RunE: func(cmd *cobra.Command, args []string) error {

		_, themesDir, _, err := resolvePaths()
		if err != nil {
			return err
		}

		themes, err := theme.ListThemes(themesDir)
		if err != nil {
			return err
		}

		for _, t := range themes {
			fmt.Println(t)
		}

		return nil
	},
}

////////////////////////////////////////////////////////////////////////////////
// PATH RESOLUTION
////////////////////////////////////////////////////////////////////////////////

func resolvePaths() (configBase, themesDir, templatesDir string, err error) {

	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", "", err
	}

	configBase = filepath.Join(home, ".config", "zenkai")
	themesDir = filepath.Join(configBase, "themes")
	templatesDir = filepath.Join(configBase, "templates")

	return
}

////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(themeCmd)

	themeCmd.AddCommand(applyCmd)
	themeCmd.AddCommand(nextCmd)
	themeCmd.AddCommand(prevCmd)
	themeCmd.AddCommand(randomCmd)
	themeCmd.AddCommand(listCmd)
}
