package screen

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/dtan4/bqc/internal/bigquery"
	"github.com/dtan4/bqc/internal/renderer"
)

var (
	textStyleDefault = tcell.StyleDefault
	textStyleSuceess = tcell.StyleDefault.Foreground(tcell.ColorGreenYellow)
	textStyleError   = tcell.StyleDefault.Foreground(tcell.ColorRed)
)

type Screen struct {
	app *tview.Application

	textArea       *tview.TextArea
	borderTextView *tview.TextView
	resultTextView *tview.TextView
	statusTextView *tview.TextView
	ctrlXTextView  *tview.TextView

	pages *tview.Pages

	bqClient *bigquery.Client
	renderer *renderer.TableRenderer

	ctrlXMode bool
}

// New creates a new TUI screen.
//
// +----------------------------------------------------------+
// | textArea                                                 |
// |                                                          |
// |                                                          |
// |                                                          |
// |                                                          |
// +----------------------------------------------------------+
// | borderTextView (height: 1)                               |
// +----------------------------------------------------------+
// | resultTextView                                           |
// |                                                          |
// |                                                          |
// |                                                          |
// |                                                          |
// +----------------------------------------------------------+
// | statusTextView                | ctrlXTextView (width: 8) |
// +----------------------------------------------------------+
func New(bqClient *bigquery.Client, renderer *renderer.TableRenderer) *Screen {
	app := tview.NewApplication()

	textArea := tview.NewTextArea().SetTextStyle(textStyleDefault)

	borderTextView := tview.NewTextView().SetText("--- result ---")
	resultTextView := tview.NewTextView().SetTextStyle(textStyleDefault).SetChangedFunc(func() {
		app.Draw()
	})
	statusTextView := tview.NewTextView().SetTextStyle(textStyleDefault).SetChangedFunc(func() {
		app.Draw()
	})
	ctrlXTextView := tview.NewTextView().SetTextStyle(textStyleDefault.Bold(true)).SetTextAlign(tview.AlignRight).SetChangedFunc(func() {
		app.Draw()
	})

	mainView := tview.NewGrid().
		SetRows(0, 1, 0, 1).
		SetColumns(0, 8).
		AddItem(textArea, 0, 0, 1, 1, 0, 0, true).
		AddItem(borderTextView, 1, 0, 1, 1, 0, 0, false).
		AddItem(resultTextView, 2, 0, 1, 1, 0, 0, false).
		AddItem(statusTextView, 3, 0, 1, 1, 0, 0, false).
		AddItem(ctrlXTextView, 3, 1, 1, 1, 0, 0, false)

	pages := tview.NewPages().AddAndSwitchToPage("main", mainView, true)

	return &Screen{
		app:            app,
		textArea:       textArea,
		borderTextView: borderTextView,
		resultTextView: resultTextView,
		statusTextView: statusTextView,
		ctrlXTextView:  ctrlXTextView,
		pages:          pages,
		bqClient:       bqClient,
		renderer:       renderer,
		ctrlXMode:      false,
	}
}

func (s *Screen) Run(ctx context.Context) error {
	s.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if s.ctrlXMode {
			s.ctrlXMode = false
			s.ctrlXTextView.SetText("")

			switch event.Key() {
			case tcell.KeyEnter:
				q := s.textArea.GetText()

				s.statusTextView.
					SetText("running query...").
					SetTextStyle(textStyleDefault)

				elapsedSecond := 1

				ticker := time.NewTicker(1 * time.Second)
				done := make(chan bool)

				go func() {
					for {
						select {
						case <-done:
							return
						case <-ticker.C:
							s.statusTextView.
								SetText(fmt.Sprintf("running query (%ds)...", elapsedSecond)).
								SetTextStyle(textStyleDefault)
							elapsedSecond += 1
						}
					}
				}()

				go func() {
					start := time.Now()
					r, err := s.bqClient.RunQuery(ctx, q)
					if err != nil {
						done <- true
						s.resultTextView.SetText(err.Error())
						s.statusTextView.
							SetText("[ERROR] cannot run query").
							SetTextStyle(textStyleError)

						return
					}

					t, err := s.renderer.Render(r)
					if err != nil {
						done <- true
						s.statusTextView.
							SetText("[ERROR] cannot render result").
							SetTextStyle(textStyleError)

						return
					}

					s.resultTextView.SetText(t)
					s.resultTextView.ScrollToBeginning()

					done <- true

					s.statusTextView.
						SetText(fmt.Sprintf("[SUCCESS] %d row(s), took %.2f seconds", len(r.Rows), time.Since(start).Seconds())).
						SetTextStyle(textStyleSuceess)
				}()

				return nil
			default:
				// do nothing
				return nil
			}
		} else {
			if event.Key() == tcell.KeyCtrlX {
				s.ctrlXMode = true
				s.ctrlXTextView.SetText("Ctrl-X")

				return nil
			}
		}

		return event
	})

	if err := s.app.SetRoot(s.pages, true).EnableMouse(true).Run(); err != nil {
		return fmt.Errorf("run TUI app: %w", err)
	}

	return nil
}
