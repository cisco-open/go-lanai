package internal

import (
	"fmt"
	"github.com/go-kit/kit/log/term"
	"io"
	"text/template"
)

// color code generation and terminal check is adopted from github.com/go-kit/kit/log/term

// ColorNames names can be used for generic color function.
// quick foreground color function with same name is also availalbe in template
var ColorNames  = map[string]Color{
	"black":   Black,
	"red":     Red,
	"green":   Green,
	"yellow":  Yellow,
	"blue":    Blue,
	"magenta": Magenta,
	"cyan":    Cyan,
	"gray":    Gray,

	"black_b":   BoldBlack,
	"red_b":     BoldRed,
	"green_b":   BoldGreen,
	"yellow_b":  BoldYellow,
	"blue_b":    BoldBlue,
	"magenta_b": BoldMagenta,
	"cyan_b":    BoldCyan,
	"gray_b":    BoldGray,
}

var (
	TmplColorFuncMap        template.FuncMap
	TmplColorFuncMapNonTerm template.FuncMap
)

type Color uint8

const (
	Default = Color(iota)

	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	Gray

	BoldBlack
	BoldRed
	BoldGreen
	BoldYellow
	BoldBlue
	BoldMagenta
	BoldCyan
	BoldGray

	numColors
)

var (
	FgColors    []string
	BgColors    []string
	ResetColor = "\x1b[39;49;22m"
)

// Implementations adopted from github.com/go-kit/kit/log/term
func init() {
	// Default
	//FgColors = append(FgColors, "\x1b[39m")
	//BgColors = append(BgColors, "\x1b[49m")
	FgColors = append(FgColors, "")
	BgColors = append(BgColors, "")

	// dark colors
	for color := Black; color < BoldBlack; color++ {
		FgColors = append(FgColors, fmt.Sprintf("\x1b[%dm", 30+color-Black))
		BgColors = append(BgColors, fmt.Sprintf("\x1b[%dm", 40+color-Black))
	}

	// bright colors
	for color := BoldBlack; color < numColors; color++ {
		FgColors = append(FgColors, fmt.Sprintf("\x1b[%d;1m", 30+color-BoldBlack))
		BgColors = append(BgColors, fmt.Sprintf("\x1b[%d;1m", 40+color-BoldBlack))
	}

	// prepare quick color function Map
	TmplColorFuncMap = template.FuncMap{"color": Colored}
	TmplColorFuncMapNonTerm = template.FuncMap{"color":NoopColored}
	for k, v := range ColorNames {
		TmplColorFuncMap[k] = MakeQuickColorFunc(v)
		TmplColorFuncMapNonTerm[k] = NoopQuickColor
	}
}

// IsTerminal returns true if w writes to a terminal.
// Implementations adopted from github.com/go-kit/kit/log/term
func IsTerminal(w io.Writer) bool {
	return term.IsTerminal(w)
}

func ColoredWithCode(s interface{}, fg, bg Color) string {
	var fgStr, bgStr string
	if fg < numColors {
		fgStr = FgColors[fg]
	}
	if bg < numColors {
		bgStr = BgColors[bg]
	}
	return fgStr + bgStr + Sprint(s) + ResetColor
}

func ColoredWithName(s interface{}, fgName, bgName string) string {
	fg, _ := ColorNames[fgName]
	bg, _ := ColorNames[bgName]
	return ColoredWithCode(s, fg, bg)
}

// Colored takes 0, 1, or 2 color names
// when present, they should be in order of fgName, bgName
func Colored(s interface{}, colorNames...string) string {
	switch len(colorNames) {
	case 0:
		return Sprint(s)
	case 1:
		return ColoredWithName(s, colorNames[0], "")
	default:
		return ColoredWithName(s, colorNames[0], colorNames[1])
	}
}

func MakeQuickColorFunc(fg Color) func(s interface{}) string {
	return func(s interface{}) string {
		return ColoredWithCode(s, fg, Default)
	}
}

func NoopColored(v interface{}, _...string) interface{} {
	return v
}

func NoopQuickColor(v interface{}) interface{} {
	return v
}


func DebugShowcase() {
	loop := func(colorFunc func(s string, c Color) string) {
		count := 0
		for k, v := range ColorNames {
			name := fmt.Sprintf("%-12s", k)
			fmt.Printf("%s ", colorFunc(name, v))
			count ++
			if count % 4 == 0 {
				fmt.Println()
			}
		}
	}

	fgColorFunc := func(s string, c Color) string {
		return ColoredWithCode(s, c, Default)
	}

	bgColorFunc := func(s string, c Color) string {
		return ColoredWithCode(s, Default, c)
	}

	// run
	loop(fgColorFunc)
	loop(bgColorFunc)
}


