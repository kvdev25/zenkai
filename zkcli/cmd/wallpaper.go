package cmd

import (
	"fmt"
	"os"

	"cli/zkcli/internal/theme"
	"cli/zkcli/internal/wallpaper"
	"github.com/spf13/cobra"
)

////////////////////////////////////////////////////////////////////////////////
// ROOT COMMAND
////////////////////////////////////////////////////////////////////////////////

var wallpaperCmd = &cobra.Command{
	Use:   "wallpaper",
	Short: "Manage wallpapers",
}

////////////////////////////////////////////////////////////////////////////////
// APPLY
////////////////////////////////////////////////////////////////////////////////

var wallpaperApplyCmd = &cobra.Command{
	Use:   "apply [file]",
	Short: "Apply a wallpaper using swww",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		if err := wallpaper.Apply(args[0]); err != nil {
			fmt.Println("Error applying wallpaper:", err)
			os.Exit(1)
		}

		fmt.Println("Wallpaper applied:", args[0])
	},
}

////////////////////////////////////////////////////////////////////////////////
// CHOOSE (File Picker via XDG Portal)
////////////////////////////////////////////////////////////////////////////////

var wallpaperChooseCmd = &cobra.Command{
	Use:   "choose",
	Short: "Choose a wallpaper using file picker",
	Run: func(cmd *cobra.Command, args []string) {

		configBase, themesDir, _, err := resolvePaths()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		currentTheme, err := theme.GetCachedTheme(configBase)
		if err != nil {
			fmt.Println("No active theme found")
			os.Exit(1)
		}

		if err := wallpaper.Choose(configBase, themesDir, currentTheme); err != nil {
			fmt.Println("Error choosing wallpaper:", err)
			os.Exit(1)
		}
	},
}

var wallpaperNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Switch to next wallpaper in current theme",
	Run: func(cmd *cobra.Command, args []string) {

		configBase, themesDir, _, err := resolvePaths()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		currentTheme, err := theme.GetCachedTheme(configBase)
		if err != nil {
			fmt.Println("No active theme found")
			os.Exit(1)
		}

		w, err := wallpaper.Next(configBase, themesDir, currentTheme)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Wallpaper:", w)
	},
}

var wallpaperPrevCmd = &cobra.Command{
	Use:   "prev",
	Short: "Switch to previous wallpaper in current theme",
	Run: func(cmd *cobra.Command, args []string) {

		configBase, themesDir, _, err := resolvePaths()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		currentTheme, err := theme.GetCachedTheme(configBase)
		if err != nil {
			fmt.Println("No active theme found")
			os.Exit(1)
		}

		w, err := wallpaper.Prev(configBase, themesDir, currentTheme)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Wallpaper:", w)
	},
}

var wallpaperRandomCmd = &cobra.Command{
	Use:   "random",
	Short: "Apply random wallpaper from current theme",
	Run: func(cmd *cobra.Command, args []string) {

		configBase, themesDir, _, err := resolvePaths()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		currentTheme, err := theme.GetCachedTheme(configBase)
		if err != nil {
			fmt.Println("No active theme found")
			os.Exit(1)
		}

		w, err := wallpaper.Random(configBase, themesDir, currentTheme)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Wallpaper:", w)
	},
}

var wallpaperRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore last wallpaper for current theme",
	Run: func(cmd *cobra.Command, args []string) {

		configBase, themesDir, _, err := resolvePaths()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		currentTheme, err := theme.GetCachedTheme(configBase)
		if err != nil {
			fmt.Println("No active theme found")
			os.Exit(1)
		}

		w, err := wallpaper.Restore(configBase, themesDir, currentTheme)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Wallpaper:", w)
	},
}

////////////////////////////////////////////////////////////////////////////////
// INIT
////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(wallpaperCmd)
	wallpaperCmd.AddCommand(wallpaperApplyCmd)
	wallpaperCmd.AddCommand(wallpaperChooseCmd)
	wallpaperCmd.AddCommand(wallpaperNextCmd)
	wallpaperCmd.AddCommand(wallpaperPrevCmd)
	wallpaperCmd.AddCommand(wallpaperRandomCmd)
	wallpaperCmd.AddCommand(wallpaperRestoreCmd)
}
