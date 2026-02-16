package wallpaper

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// APPLY (existing)
////////////////////////////////////////////////////////////////////////////////

// Apply executes:
//
// swww img <path>
func Apply(path string) error {

	expanded := expandHome(path)

	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return errors.New("provided path is a directory")
	}

	cmd := exec.Command("swww", "img", absPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Executing: swww img", absPath)

	return cmd.Run()
}

////////////////////////////////////////////////////////////////////////////////
// CHOOSE (XDG Portal File Picker)
////////////////////////////////////////////////////////////////////////////////

// Choose opens a graphical file picker using:
//
// org.freedesktop.portal.FileChooser
//
// After user selects a file, it calls Apply().

func Choose(configBase, themesDir, currentTheme string) error {

	startDir := filepath.Join(themesDir, currentTheme, "backgrounds")

	// Ensure directory exists
	if _, err := os.Stat(startDir); err != nil {
		return fmt.Errorf("backgrounds folder not found for theme: %s", currentTheme)
	}

	// Launch zenity file picker
	cmd := exec.Command(
		"zenity",
		"--file-selection",
		"--title=Select Wallpaper",
		"--filename="+startDir+"/",
	)

	out, err := cmd.Output()
	if err != nil {
		// User cancels → zenity exits with non-zero
		return fmt.Errorf("selection cancelled")
	}

	filePath := strings.TrimSpace(string(out))
	if filePath == "" {
		return fmt.Errorf("no file selected")
	}

	fmt.Println("Selected:", filePath)

	// Cache wallpaper for theme
	if err := cacheWallpaper(configBase, currentTheme, filePath); err != nil {
		return err
	}

	// Apply wallpaper
	return Apply(filePath)
}

////////////////////////////////////////////////////////////////////////////////
// UTIL
////////////////////////////////////////////////////////////////////////////////

func expandHome(path string) string {
	if len(path) > 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

////////////////////////////////////////////////////////////////////////////////
// THEME-AWARE WALLPAPER LOGIC
////////////////////////////////////////////////////////////////////////////////

const wallpaperStateDir = ".wallpaper_state"

func getWallpaperStateDir(configBase string) string {
	return filepath.Join(configBase, wallpaperStateDir)
}

func cacheWallpaper(configBase, themeName, wallpaperPath string) error {
	dir := getWallpaperStateDir(configBase)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	stateFile := filepath.Join(dir, themeName)
	return os.WriteFile(stateFile, []byte(wallpaperPath), 0644)
}

func getCachedWallpaper(configBase, themeName string) (string, error) {
	stateFile := filepath.Join(getWallpaperStateDir(configBase), themeName)

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func getWallpapers(themesDir, themeName string) ([]string, error) {
	bgDir := filepath.Join(themesDir, themeName, "backgrounds")

	entries, err := os.ReadDir(bgDir)
	if err != nil {
		return nil, err
	}

	var wallpapers []string

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		wallpapers = append(wallpapers, filepath.Join(bgDir, e.Name()))
	}

	if len(wallpapers) == 0 {
		return nil, errors.New("no wallpapers found for theme")
	}

	return wallpapers, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS
////////////////////////////////////////////////////////////////////////////////

func Next(configBase, themesDir, currentTheme string) (string, error) {

	wallpapers, err := getWallpapers(themesDir, currentTheme)
	if err != nil {
		return "", err
	}

	current, _ := getCachedWallpaper(configBase, currentTheme)

	index := 0
	for i, w := range wallpapers {
		if w == current {
			index = (i + 1) % len(wallpapers)
			break
		}
	}

	next := wallpapers[index]

	if err := Apply(next); err != nil {
		return "", err
	}

	if err := cacheWallpaper(configBase, currentTheme, next); err != nil {
		return "", err
	}

	return next, nil
}

func Prev(configBase, themesDir, currentTheme string) (string, error) {

	wallpapers, err := getWallpapers(themesDir, currentTheme)
	if err != nil {
		return "", err
	}

	current, _ := getCachedWallpaper(configBase, currentTheme)

	index := 0
	for i, w := range wallpapers {
		if w == current {
			index = (i - 1 + len(wallpapers)) % len(wallpapers)
			break
		}
	}

	prev := wallpapers[index]

	if err := Apply(prev); err != nil {
		return "", err
	}

	if err := cacheWallpaper(configBase, currentTheme, prev); err != nil {
		return "", err
	}

	return prev, nil
}

func Random(configBase, themesDir, currentTheme string) (string, error) {

	wallpapers, err := getWallpapers(themesDir, currentTheme)
	if err != nil {
		return "", err
	}

	randIndex := time.Now().UnixNano() % int64(len(wallpapers))
	random := wallpapers[randIndex]

	if err := Apply(random); err != nil {
		return "", err
	}

	if err := cacheWallpaper(configBase, currentTheme, random); err != nil {
		return "", err
	}

	return random, nil
}

func Restore(configBase, themesDir, currentTheme string) (string, error) {

	wallpaper, err := getCachedWallpaper(configBase, currentTheme)
	if err == nil {
		if err := Apply(wallpaper); err == nil {
			return wallpaper, nil
		}
	}

	// fallback to random
	return Random(configBase, themesDir, currentTheme)
}
