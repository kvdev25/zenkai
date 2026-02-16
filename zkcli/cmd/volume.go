package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Manage Volume",
}

var volumeSetCmd = &cobra.Command{
	Use:   "set [value]",
	Short: "Set system volume and show notification",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value := args[0]

		//  Set volume
		setCmd := exec.Command(
			"wpctl",
			"set-volume",
			"-l", "2",
			"@DEFAULT_AUDIO_SINK@",
			value,
		)

		if err := setCmd.Run(); err != nil {
			return fmt.Errorf("failed to set volume: %w", err)
		}

		//  Get current volume
		getCmd := exec.Command(
			"wpctl",
			"get-volume",
			"@DEFAULT_AUDIO_SINK@",
		)

		out, err := getCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get volume: %w", err)
		}

		// wpctl output example:
		// Volume: 0.53
		fields := strings.Fields(string(out))
		if len(fields) < 2 {
			return fmt.Errorf("unexpected wpctl output")
		}

		// Convert to percent
		volFloat := fields[1]
		var percent int
		fmt.Sscanf(volFloat, "%f", &volFloat) // this is wrong, so we’ll fix below

		// Proper parse:
		var volValue float64
		fmt.Sscanf(volFloat, "%f", &volValue)
		percent = int(volValue * 100)

		//  Send notification
		notifyCmd := exec.Command(
			"notify-send",
			"-t", "1500",
			"-h", fmt.Sprintf("int:value:%d", percent),
			"-h", "string:x-canonical-private-synchronous:volume",
			"Volume",
			fmt.Sprintf("%d%%", percent),
		)

		if err := notifyCmd.Run(); err != nil {
			return fmt.Errorf("failed to send notification: %w", err)
		}

		return nil
	},
}

var volumeMuteCmd = &cobra.Command{
	Use:   "mute",
	Short: "Toggle mute and show notification",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Toggle mute
		toggleCmd := exec.Command(
			"wpctl",
			"set-mute",
			"@DEFAULT_AUDIO_SINK@",
			"toggle",
		)

		if err := toggleCmd.Run(); err != nil {
			return fmt.Errorf("failed to toggle mute: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(volumeCmd)
	volumeCmd.AddCommand(volumeSetCmd)
	volumeCmd.AddCommand(volumeMuteCmd)
}
