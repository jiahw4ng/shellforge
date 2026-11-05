package ui

// Palette holds the ANSI 256-color codes one Theme uses. Codes are strings
// because both Lipgloss and Glamour take them in that form.
type Palette struct {
	// Accent colors headings, controls, and the selected item.
	Accent string
	// Text colors explanatory text.
	Text string
	// Muted colors secondary information.
	Muted string
	// Primary colors the application frame and emphasized text.
	Primary string
	Red     string
	Green   string
	Orange  string
	// InlineCode and InlineCodeBackground color inline commands in lesson text.
	InlineCode           string
	InlineCodeBackground string
}

// Theme is one learner-selectable color scheme. ID is the stable value
// persisted in learner settings.
type Theme struct {
	ID    string
	Label string
	// LightBackground marks palettes designed for light terminal backgrounds,
	// so lesson Markdown starts from a matching base style.
	LightBackground bool
	Palette         Palette
}

// Themes lists every theme the learner can choose. The first entry is the default.
var Themes = []Theme{
	{
		ID:    "dark",
		Label: "Dark",
		Palette: Palette{
			Accent: "33", Text: "252", Muted: "250", Primary: "15",
			Red: "196", Green: "42", Orange: "208",
			InlineCode: "222", InlineCodeBackground: "238",
		},
	},
	{
		ID:              "light",
		Label:           "Light",
		LightBackground: true,
		Palette: Palette{
			Accent: "25", Text: "235", Muted: "240", Primary: "16",
			Red: "160", Green: "28", Orange: "130",
			InlineCode: "88", InlineCodeBackground: "254",
		},
	},
	{
		// High contrast assumes a dark terminal background.
		ID:    "high-contrast",
		Label: "High contrast",
		Palette: Palette{
			Accent: "51", Text: "231", Muted: "253", Primary: "231",
			Red: "203", Green: "46", Orange: "214",
			InlineCode: "231", InlineCodeBackground: "238",
		},
	},
}

// DefaultTheme is used until the learner chooses or loads another theme.
var DefaultTheme = Themes[0]

// ThemeByID returns the theme with the given persisted ID.
func ThemeByID(id string) (Theme, bool) {
	for _, theme := range Themes {
		if theme.ID == id {
			return theme, true
		}
	}
	return Theme{}, false
}

var activeTheme = DefaultTheme

// CurrentTheme returns the theme most recently applied with ApplyTheme.
func CurrentTheme() Theme {
	return activeTheme
}

// ApplyTheme rebuilds every shared color and style from theme. Call it from
// the Bubble Tea update loop, never from a background command, because
// rendering reads the styles from View.
func ApplyTheme(theme Theme) {
	activeTheme = theme
	applyPalette(theme.Palette)
}

func init() {
	ApplyTheme(DefaultTheme)
}
