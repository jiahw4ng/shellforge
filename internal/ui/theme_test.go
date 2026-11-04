package ui

import (
	"image/color"
	"math"
	"strconv"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestThemesHaveUniqueIDsAndDefaultFirst(t *testing.T) {
	seen := map[string]bool{}
	for _, theme := range Themes {
		if theme.ID == "" || theme.Label == "" {
			t.Errorf("theme %#v has an empty ID or label", theme)
		}
		if seen[theme.ID] {
			t.Errorf("theme ID %q is not unique", theme.ID)
		}
		seen[theme.ID] = true
	}
	if DefaultTheme.ID != Themes[0].ID || DefaultTheme.ID != "dark" {
		t.Fatalf("DefaultTheme = %q, want the first theme and dark", DefaultTheme.ID)
	}
}

func TestThemeByIDRejectsUnknownID(t *testing.T) {
	if theme, ok := ThemeByID("light"); !ok || theme.ID != "light" {
		t.Fatalf("ThemeByID(light) = %#v, %v; want light", theme, ok)
	}
	if theme, ok := ThemeByID("solarized"); ok {
		t.Fatalf("ThemeByID(solarized) = %#v, true; want not found", theme)
	}
}

func TestApplyThemeRebuildsSharedColorsAndFrame(t *testing.T) {
	t.Cleanup(func() { ApplyTheme(DefaultTheme) })

	light, _ := ThemeByID("light")
	ApplyTheme(light)

	if CurrentTheme().ID != "light" {
		t.Fatalf("CurrentTheme() = %q, want light", CurrentTheme().ID)
	}
	if !sameColor(PrimaryColor, parsedColor(t, light.Palette.Primary)) {
		t.Errorf("PrimaryColor does not follow the light palette")
	}
	if !sameColor(AccentColor, parsedColor(t, light.Palette.Accent)) {
		t.Errorf("AccentColor does not follow the light palette")
	}
	if !sameColor(ApplicationFrameStyle.GetBorderTopForeground(), PrimaryColor) {
		t.Errorf("ApplicationFrameStyle border does not follow the primary color")
	}
	if !sameColor(TitleStyle.GetForeground(), AccentColor) {
		t.Errorf("TitleStyle does not follow the accent color")
	}

	ApplyTheme(DefaultTheme)
	if !sameColor(AccentColor, parsedColor(t, DefaultTheme.Palette.Accent)) {
		t.Errorf("AccentColor did not return to the dark palette")
	}
}

// Contrast is checked against the background each palette is designed for:
// black for dark and high contrast, white for light. The 256-color cube and
// grayscale ramp are fixed by xterm, so the ratios are deterministic.
func TestThemePalettesMeetContrastTargets(t *testing.T) {
	for _, theme := range Themes {
		t.Run(theme.ID, func(t *testing.T) {
			background := rgb{0, 0, 0}
			if theme.LightBackground {
				background = rgb{255, 255, 255}
			}
			palette := theme.Palette

			textMinimum := 7.0
			if theme.LightBackground {
				textMinimum = 4.5
			}
			for name, test := range map[string]struct {
				code    string
				minimum float64
			}{
				"text":    {palette.Text, textMinimum},
				"muted":   {palette.Muted, textMinimum},
				"primary": {palette.Primary, textMinimum},
				"accent":  {palette.Accent, 4.5},
				"red":     {palette.Red, 4.5},
				"green":   {palette.Green, 4.5},
				"orange":  {palette.Orange, 4.5},
			} {
				if ratio := contrastRatio(xtermColor(t, test.code), background); ratio < test.minimum {
					t.Errorf("%s %s contrast = %.2f:1, want at least %.1f:1", theme.ID, name, ratio, test.minimum)
				}
			}

			codeBackground := xtermColor(t, palette.InlineCodeBackground)
			if ratio := contrastRatio(xtermColor(t, palette.InlineCode), codeBackground); ratio < 4.5 {
				t.Errorf("%s inline code contrast = %.2f:1, want at least 4.5:1", theme.ID, ratio)
			}
		})
	}
}

func parsedColor(t *testing.T, code string) color.Color {
	t.Helper()
	return lipgloss.Color(code)
}

func sameColor(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

type rgb struct{ r, g, b int }

// xtermColor converts an ANSI 256-color code to RGB. Codes 0-14 are terminal
// defined and so cannot be checked; they are rejected rather than guessed.
func xtermColor(t *testing.T, code string) rgb {
	t.Helper()
	n, err := strconv.Atoi(code)
	if err != nil || n < 0 || n > 255 {
		t.Fatalf("color code %q is not an ANSI 256-color code", code)
	}
	switch {
	case n == 15:
		return rgb{255, 255, 255}
	case n < 16:
		t.Fatalf("color code %q is a terminal-defined system color; use 15 or 16-255", code)
	case n < 232:
		levels := [6]int{0, 95, 135, 175, 215, 255}
		n -= 16
		return rgb{levels[n/36], levels[n/6%6], levels[n%6]}
	}
	gray := 8 + 10*(n-232)
	return rgb{gray, gray, gray}
}

func contrastRatio(a, b rgb) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func luminance(c rgb) float64 {
	channel := func(v int) float64 {
		s := float64(v) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(c.r) + 0.7152*channel(c.g) + 0.0722*channel(c.b)
}
