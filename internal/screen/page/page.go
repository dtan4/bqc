package page

import "github.com/rivo/tview"

type Page interface {
	tview.Primitive

	Init() error
	Close() error
}
