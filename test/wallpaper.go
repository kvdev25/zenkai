package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/godbus/dbus/v5"
)

func main() {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	home, _ := os.UserHomeDir()
	startDir := filepath.Join(home, "pictures", "wallpapers")

	obj := conn.Object(
		"org.freedesktop.portal.Desktop",
		"/org/freedesktop/portal/desktop",
	)

	// Options map
	options := map[string]dbus.Variant{
		"current_folder": dbus.MakeVariant("file://" + startDir),
	}

	// Call OpenFile
	var requestPath dbus.ObjectPath
	err = obj.Call(
		"org.freedesktop.portal.FileChooser.OpenFile",
		0,
		"", // parent window
		"Select Wallpaper",
		options,
	).Store(&requestPath)

	if err != nil {
		panic(err)
	}

	// Listen for response signal
	rule := fmt.Sprintf(
		"type='signal',interface='org.freedesktop.portal.Request',path='%s'",
		requestPath,
	)

	err = conn.BusObject().Call(
		"org.freedesktop.DBus.AddMatch",
		0,
		rule,
	).Err

	if err != nil {
		panic(err)
	}

	signals := make(chan *dbus.Signal, 1)
	conn.Signal(signals)

	for signal := range signals {
		if signal.Path == requestPath &&
			signal.Name == "org.freedesktop.portal.Request.Response" {

			responseCode := signal.Body[0].(uint32)
			results := signal.Body[1].(map[string]dbus.Variant)

			if responseCode != 0 {
				fmt.Println("Selection cancelled")
				return
			}

			uris := results["uris"].Value().([]string)
			if len(uris) == 0 {
				fmt.Println("No file selected")
				return
			}

			parsed, err := url.Parse(uris[0])
			if err != nil {
				panic(err)
			}

			filePath := parsed.Path
			fmt.Println("Selected:", filePath)

			runSwww(filePath)
			return
		}
	}
}

func runSwww(path string) {
	cmd := exec.Command("swww", "img", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed to run swww:", err)
	}
}
