package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var brightnessCmd = &cobra.Command{
	Use:   "brightness",
	Short: "Manage screen brightness",
}

var brightnessSetCmd = &cobra.Command{
	Use:   "set [value]",
	Short: "Set brightness and show notification",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value := args[0]

		// Set brightness
		setCmd := exec.Command("brightnessctl", "set", value)
		if err := setCmd.Run(); err != nil {
			return fmt.Errorf("failed to set brightness: %w", err)
		}

		// Get current brightness (percentage)
		getCmd := exec.Command("brightnessctl", "-q", "get")
		out, err := getCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get brightness: %w", err)
		}

		brightness := strings.TrimSpace(string(out))

		// Send notification
		notifyCmd := exec.Command(
			"notify-send",
			"-t", "1500",
			"-h", fmt.Sprintf("int:value:%s", brightness),
			"-h", "string:x-canonical-private-synchronous:brightness",
			"Brightness",
			fmt.Sprintf("%s%%", brightness),
		)

		if err := notifyCmd.Run(); err != nil {
			return fmt.Errorf("failed to send notification: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(brightnessCmd)
	brightnessCmd.AddCommand(brightnessSetCmd)
}
