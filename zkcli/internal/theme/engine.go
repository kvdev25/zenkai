package theme

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

////////////////////////////////////////////////////////////////////////////////
// TEMPLATE REGEX (chained modifiers + arguments)
////////////////////////////////////////////////////////////////////////////////

var templateRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.]+)((?::[a-zA-Z0-9_]+(?:\([^)]*\))?)*)\s*\}\}`)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC API
////////////////////////////////////////////////////////////////////////////////

func ParseNestedToml(path string) (map[string]string, error) {
	var raw map[string]any
	if _, err := toml.DecodeFile(path, &raw); err != nil {
		return nil, err
	}

	values := make(map[string]string)
	flattenMap("", raw, values)
	return values, nil
}

func ApplyThemeParallel(templateDir string, values map[string]string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 16)

	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			if err := processTemplate(p, values); err != nil {
				errChan <- err
			}
		}(path)

		return nil
	})

	if err != nil {
		return err
	}

	wg.Wait()
	close(errChan)

	for e := range errChan {
		return e
	}

	return applySystemMode(values)
}

////////////////////////////////////////////////////////////////////////////////
// APPLY BY NAME (NEW — DOES NOT MODIFY ENGINE)
////////////////////////////////////////////////////////////////////////////////

func ApplyThemeByName(configBase, themesDir, templatesDir, themeName string) error {
	themePath := filepath.Join(themesDir, themeName, "config.toml")

	values, err := ParseNestedToml(themePath)
	if err != nil {
		return err
	}

	err = ApplyThemeParallel(templatesDir, values)

	// Always try to cache theme if parsing worked
	cacheErr := CacheCurrentTheme(configBase, themeName)

	if err != nil {
		return err
	}

	return cacheErr
}

////////////////////////////////////////////////////////////////////////////////
// TEMPLATE PROCESSING
////////////////////////////////////////////////////////////////////////////////

func processTemplate(templatePath string, values map[string]string) error {
	file, err := os.Open(templatePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	firstLine, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}

	firstLine = strings.TrimSpace(firstLine)

	var paths []string
	var command string

	if strings.Contains(firstLine, "::") {
		parts := strings.SplitN(firstLine, "::", 2)
		if strings.TrimSpace(parts[0]) != "" {
			paths = strings.Fields(strings.TrimSpace(parts[0]))
		}
		command = strings.TrimSpace(parts[1])
	} else if firstLine != "" {
		paths = strings.Fields(firstLine)
	}

	rest, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	rendered := renderTemplate(string(rest), values)

	for _, p := range paths {
		dest := expandHome(p)

		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}

		if err := atomicWrite(dest, []byte(rendered)); err != nil {
			return err
		}

		fmt.Println("Written:", dest)
	}

	if command != "" {
		cmd := exec.Command("sh", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// TEMPLATE RENDERING
////////////////////////////////////////////////////////////////////////////////

func renderTemplate(content string, values map[string]string) string {
	return templateRegex.ReplaceAllStringFunc(content, func(match string) string {
		parts := templateRegex.FindStringSubmatch(match)
		key := parts[1]
		modifierChain := parts[2]

		value, ok := values[key]
		if !ok {
			return match
		}

		if modifierChain != "" {
			modifiers := parseModifiers(modifierChain)
			for _, m := range modifiers {
				value = applyModifier(value, m.name, m.args)
			}
		}

		return value
	})
}

////////////////////////////////////////////////////////////////////////////////
// MODIFIER PARSING
////////////////////////////////////////////////////////////////////////////////

type modifier struct {
	name string
	args []string
}

func parseModifiers(chain string) []modifier {
	chain = strings.TrimPrefix(chain, ":")
	rawMods := strings.Split(chain, ":")

	var mods []modifier

	for _, raw := range rawMods {
		if strings.Contains(raw, "(") {
			name := raw[:strings.Index(raw, "(")]
			argStr := raw[strings.Index(raw, "(")+1 : len(raw)-1]

			var args []string
			if argStr != "" {
				args = strings.Split(argStr, ",")
			}

			mods = append(mods, modifier{name, args})
		} else {
			mods = append(mods, modifier{raw, nil})
		}
	}

	return mods
}

////////////////////////////////////////////////////////////////////////////////
// MODIFIERS
////////////////////////////////////////////////////////////////////////////////

func applyModifier(value, name string, args []string) string {
	switch name {

	case "strip":
		return strings.TrimPrefix(value, "#")

	case "upper":
		return strings.ToUpper(value)

	case "lower":
		return strings.ToLower(value)

	case "replace":
		if len(args) == 2 {
			return strings.ReplaceAll(value, args[0], args[1])
		}

	case "repeat":
		if len(args) == 1 {
			n, _ := strconv.Atoi(args[0])
			return strings.Repeat(value, n)
		}

	case "lighten":
		if len(args) == 1 {
			p, _ := strconv.ParseFloat(args[0], 64)
			return adjustLightness(value, p)
		}

	case "darken":
		if len(args) == 1 {
			p, _ := strconv.ParseFloat(args[0], 64)
			return adjustLightness(value, -p)
		}

	case "saturate":
		if len(args) == 1 {
			p, _ := strconv.ParseFloat(args[0], 64)
			return adjustSaturation(value, p)
		}

	case "desaturate":
		if len(args) == 1 {
			p, _ := strconv.ParseFloat(args[0], 64)
			return adjustSaturation(value, -p)
		}

	case "hue":
		if len(args) == 1 {
			deg, _ := strconv.ParseFloat(args[0], 64)
			return adjustHue(value, deg)
		}

	case "hsl":
		h, s, l := hexToHSL(value)
		return fmt.Sprintf("%.0f,%.0f%%,%.0f%%", h, s*100, l*100)

	case "rgb":
		r, g, b := hexToRGB(value)
		return fmt.Sprintf("%d,%d,%d", r, g, b)

	case "r":
		r, _, _ := hexToRGB(value)
		return strconv.Itoa(r)

	case "g":
		_, g, _ := hexToRGB(value)
		return strconv.Itoa(g)

	case "b":
		_, _, b := hexToRGB(value)
		return strconv.Itoa(b)
	}

	return value
}

////////////////////////////////////////////////////////////////////////////////
// COLOR UTILITIES
////////////////////////////////////////////////////////////////////////////////

func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0
	}

	r, _ := strconv.ParseInt(hex[0:2], 16, 0)
	g, _ := strconv.ParseInt(hex[2:4], 16, 0)
	b, _ := strconv.ParseInt(hex[4:6], 16, 0)

	return int(r), int(g), int(b)
}

func hexToHSL(hex string) (float64, float64, float64) {
	r, g, b := hexToRGB(hex)

	rf := float64(r) / 255
	gf := float64(g) / 255
	bf := float64(b) / 255

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))

	l := (max + min) / 2

	var h, s float64

	if max == min {
		h = 0
		s = 0
	} else {
		d := max - min

		if l > 0.5 {
			s = d / (2 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case rf:
			h = (gf - bf) / d
			if gf < bf {
				h += 6
			}
		case gf:
			h = (bf-rf)/d + 2
		case bf:
			h = (rf-gf)/d + 4
		}

		h *= 60
	}

	return h, s, l
}

func hslToHex(h, s, l float64) string {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	var r, g, b float64

	if s == 0 {
		r = l
		g = l
		b = l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}

		p := 2*l - q
		hk := h / 360

		r = hueToRGB(p, q, hk+1.0/3.0)
		g = hueToRGB(p, q, hk)
		b = hueToRGB(p, q, hk-1.0/3.0)
	}

	return fmt.Sprintf("#%02x%02x%02x",
		int(r*255),
		int(g*255),
		int(b*255),
	)
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

func adjustLightness(hex string, percent float64) string {
	h, s, l := hexToHSL(hex)
	l += percent / 100
	l = math.Max(0, math.Min(1, l))
	return hslToHex(h, s, l)
}

func adjustSaturation(hex string, percent float64) string {
	h, s, l := hexToHSL(hex)
	s += percent / 100
	s = math.Max(0, math.Min(1, s))
	return hslToHex(h, s, l)
}

func adjustHue(hex string, degrees float64) string {
	h, s, l := hexToHSL(hex)
	h += degrees
	return hslToHex(h, s, l)
}

////////////////////////////////////////////////////////////////////////////////
// SYSTEM MODE
////////////////////////////////////////////////////////////////////////////////

func applySystemMode(values map[string]string) error {
	mode, ok := values["mode"]
	if !ok {
		return nil
	}

	var cmd string

	switch strings.ToLower(mode) {
	case "dark":
		cmd = `gsettings set org.gnome.desktop.interface color-scheme "prefer-dark";
		       gsettings set org.gnome.desktop.interface gtk-theme "Adwaita-dark"`
	case "light":
		cmd = `gsettings set org.gnome.desktop.interface color-scheme "prefer-light";
		       gsettings set org.gnome.desktop.interface gtk-theme "Adwaita"`
	}

	if cmd == "" {
		return nil
	}

	command := exec.Command("sh", "-c", cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	fmt.Println("Applying system mode:", mode)
	return command.Run()
}

////////////////////////////////////////////////////////////////////////////////
// ATOMIC WRITE
////////////////////////////////////////////////////////////////////////////////

func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)

	tmpFile, err := os.CreateTemp(dir, "tmp-*")
	if err != nil {
		return err
	}

	tmpName := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpName, path)
}

////////////////////////////////////////////////////////////////////////////////
// TOML FLATTENER
////////////////////////////////////////////////////////////////////////////////

func flattenMap(prefix string, input map[string]any, output map[string]string) {
	for key, value := range input {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]any:
			flattenMap(fullKey, v, output)
		default:
			output[fullKey] = fmt.Sprintf("%v", v)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// UTIL
////////////////////////////////////////////////////////////////////////////////

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

////////////////////////////////////////////////////////////////////////////////
// THEME STATE (NEW FEATURE ONLY)
////////////////////////////////////////////////////////////////////////////////

const stateFileName = ".theme_state"

func CacheCurrentTheme(configBase, themeName string) error {
	statePath := filepath.Join(configBase, stateFileName)
	return os.WriteFile(statePath, []byte(themeName), 0644)
}

func GetCachedTheme(configBase string) (string, error) {
	statePath := filepath.Join(configBase, stateFileName)
	data, err := os.ReadFile(statePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func ListThemes(themesDir string) ([]string, error) {
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return nil, err
	}

	var themes []string
	for _, e := range entries {
		if e.IsDir() {
			themes = append(themes, e.Name())
		}
	}

	sort.Strings(themes) // ✅ CRITICAL FIX

	return themes, nil
}

func Choose(configBase, themesDir string) (string, error) {
	percentSign := "%"
	cmdStr := fmt.Sprintf(
		"zkcli theme list | fzf --preview-window=right,70%s --preview='fzf-preview %s/{1}/preview.png' | xargs -xo zkcli theme apply",
		percentSign, themesDir,
	)

	cmd := exec.Command("xdg-terminal-exec", "bash", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run theme choose pipeline: %w", err)
	}

	// Optional: read the applied theme from cache
	currentTheme, err := GetCachedTheme(configBase)
	if err != nil {
		return "", fmt.Errorf("failed to get cached theme: %w", err)
	}

	return currentTheme, nil
}

func GetNextTheme(configBase, themesDir string) (string, error) {
	themes, err := ListThemes(themesDir)
	if err != nil {
		return "", err
	}
	if len(themes) == 0 {
		return "", fmt.Errorf("no themes found")
	}

	current, _ := GetCachedTheme(configBase)

	for i, t := range themes {
		if t == current {
			return themes[(i+1)%len(themes)], nil
		}
	}

	return themes[0], nil
}

func GetPrevTheme(configBase, themesDir string) (string, error) {
	themes, err := ListThemes(themesDir)
	if err != nil {
		return "", err
	}
	if len(themes) == 0 {
		return "", fmt.Errorf("no themes found")
	}

	current, _ := GetCachedTheme(configBase)

	for i, t := range themes {
		if t == current {
			return themes[(i-1+len(themes))%len(themes)], nil
		}
	}

	return themes[0], nil
}

func GetRandomTheme(themesDir string) (string, error) {
	themes, err := ListThemes(themesDir)
	if err != nil {
		return "", err
	}
	if len(themes) == 0 {
		return "", fmt.Errorf("no themes found")
	}

	rand.Seed(time.Now().UnixNano())
	return themes[rand.Intn(len(themes))], nil
}
