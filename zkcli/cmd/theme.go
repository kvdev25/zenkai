package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"cli/zkcli/internal/theme"
	"github.com/spf13/cobra"
)

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage themes",
}

var themeApplyCmd = &cobra.Command{
	Use:   "apply [name]",
	Short: "Apply a theme",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		themeName := args[0]

		configBase, err := getConfigDir()
		if err != nil {
			fmt.Println("Error resolving config directory:", err)
			os.Exit(1)
		}

		themesDir := filepath.Join(configBase, "themes")
		templatesDir := filepath.Join(configBase, "templates")

		configPath := filepath.Join(themesDir, themeName, "config.toml")

		values, err := theme.ParseNestedToml(configPath)
		if err != nil {
			fmt.Println("Error loading theme:", err)
			os.Exit(1)
		}

		if err := theme.ApplyThemeParallel(templatesDir, values); err != nil {
			fmt.Println("Error applying theme:", err)
			os.Exit(1)
		}

		fmt.Println("Theme applied:", themeName)
	},
}

func init() {
	rootCmd.AddCommand(themeCmd)
	themeCmd.AddCommand(themeApplyCmd)
}

func getConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "zenkai"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "zenkai"), nil
}
