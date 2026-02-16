package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zkcli",
	Short: "Zenkai Hyprland CLI",
	Long:  "zkcli manages themes, wallpapers and configuration for zenkai Hyprland dots.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
