package page

import "github.com/gdamore/tcell/v2"

var (
	textStyleDefault = tcell.StyleDefault
	textStyleSuceess = tcell.StyleDefault.Foreground(tcell.ColorGreenYellow)
	textStyleError   = tcell.StyleDefault.Foreground(tcell.ColorRed)
)
