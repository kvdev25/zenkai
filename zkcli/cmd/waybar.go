package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

var waybarCmd = &cobra.Command{
	Use:   "waybar",
	Short: "Manage Waybar",
	Long:  "change waybar themes,",
}

var waybarListCmd = &cobra.Command{
	Use:   "list",
	Short: "List waybar themes",
	RunE: func(cmd *cobra.Command, args []string) error {

		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		configBase := filepath.Join(home, ".config", "waybar")
		entries, err := os.ReadDir(configBase)

		var themes []string
		for _, e := range entries {
			if e.IsDir() {
				themes = append(themes, e.Name())
			}
		}

		sort.Strings(themes)

		for _, t := range themes {
			fmt.Println(t)
		}

		return nil
	},
}

var waybarApplyCmd = &cobra.Command{
	Use:   "apply [theme]",
	Short: "Set Waybar Theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if args[0] == "" || args[0] == " " {
			return fmt.Errorf("Provide a valid waybar theme")
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		configBase := filepath.Join(home, ".config", "waybar")
		themeDir := filepath.Join(configBase, args[0])

		newJsonc := filepath.Join(configBase, "config.jsonc")
		newCss := filepath.Join(configBase, "style.css")
		oldJsonc := filepath.Join(themeDir, "config.jsonc")
		oldCss := filepath.Join(themeDir, "style.css")

		if err := os.Remove(newJsonc); err != nil {
			return err
		}
		if err := os.Remove(newCss); err != nil {
			return err
		}

		if err := os.Symlink(oldJsonc, newJsonc); err != nil {
			return err
		}
		if err := os.Symlink(oldCss, newCss); err != nil {
			return err
		}

		reloadCommand := exec.Command("pkill", "-SIGUSR2", "waybar")

		if err := reloadCommand.Run(); err != nil {
			return fmt.Errorf("failed to set volume: %w", err)
		}

		return nil
	},
}

var waybarReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload Waybar",
	RunE: func(cmd *cobra.Command, args []string) error {

		reloadCommand := exec.Command("pkill", "-SIGUSR2", "waybar")

		if err := reloadCommand.Run(); err != nil {
			return fmt.Errorf("failed to set volume: %w", err)
		}

		return nil
	},
}

var waybarToggleCmd = &cobra.Command{
	Use:   "toggle",
	Short: "Toggle Waybar",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := exec.Command("pkill", "waybar").Run()

		if err != nil {
			exec.Command("waybar").Start()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(waybarCmd)
	waybarCmd.AddCommand(waybarListCmd)
	waybarCmd.AddCommand(waybarApplyCmd)
	waybarCmd.AddCommand(waybarReloadCmd)
	waybarCmd.AddCommand(waybarToggleCmd)
}
