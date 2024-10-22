package screen

import (
	"context"
	"fmt"

	"github.com/rivo/tview"

	"github.com/dtan4/bqc/internal/bigquery"
	"github.com/dtan4/bqc/internal/checkpoint"
	"github.com/dtan4/bqc/internal/history"
	"github.com/dtan4/bqc/internal/screen/page"
)

const (
	pageNameQuery = "query"
)

type Screen struct {
	app *tview.Application

	pages map[string]page.Page

	bqClient   *bigquery.Client
	checkpoint *checkpoint.Checkpoint
	history    history.Storage
}

func New(
	bqClient *bigquery.Client,
	checkpoint *checkpoint.Checkpoint,
	history history.Storage,
) *Screen {
	app := tview.NewApplication()

	pages := map[string]page.Page{
		pageNameQuery: page.NewQuery(app, bqClient, checkpoint, history),
	}

	return &Screen{
		app:        app,
		pages:      pages,
		bqClient:   bqClient,
		checkpoint: checkpoint,
		history:    history,
	}
}

func (s *Screen) Run(ctx context.Context) error {
	ps := tview.NewPages()

	for k, p := range s.pages {
		p.Init()
		ps.AddPage(k, p, true, false)
	}
	defer func() {
		for _, p := range s.pages {
			_ = p.Close()
		}
	}()

	ps.SwitchToPage(pageNameQuery)

	if err := s.app.SetRoot(ps, true).EnableMouse(true).Run(); err != nil {
		return fmt.Errorf("run TUI app: %w", err)
	}

	return nil
}
