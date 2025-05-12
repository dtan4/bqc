package page

import (
	"context"
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/dtan4/bqc/internal/bigquery"
	"github.com/dtan4/bqc/internal/checkpoint"
	"github.com/dtan4/bqc/internal/history"
	"github.com/dtan4/bqc/internal/renderer"
)

type Query struct {
	*tview.Grid

	app *tview.Application

	bqClient         *bigquery.Client
	defaultRenderer  renderer.Renderer
	markdownRenderer *renderer.MarkdownRenderer
	tsvRenderer      *renderer.TSVRenderer
	checkpoint       *checkpoint.Checkpoint
	history          history.Storage

	textArea          *tview.TextArea
	borderTextView    *tview.TextView
	resultTextView    *tview.TextView
	statusTextView    *tview.TextView
	ctrlXTextView     *tview.TextView
	cursorPosTextView *tview.TextView

	ctrlXMode  bool
	lastResult *bigquery.Result
}

var _ Page = (*Query)(nil)

// NewQuery creates Query page.
//
// +-------------------------------------------------------------------+
// | textArea                                                          |
// |                                                                   |
// |                                                                   |
// |                                                                   |
// |                                                                   |
// +-------------------------------------------------------------------+
// | borderTextView (height: 1)                                        |
// +-------------------------------------------------------------------+
// | resultTextView                                                    |
// |                                                                   |
// |                                                                   |
// |                                                                   |
// |                                                                   |
// +-------------------------------------------------------------------+
// | statusTextView                | ctrlXTextView | cursorPosTextView |
// |                               | (width: 8)    | (width: 18)       |
// +-------------------------------------------------------------------+
func NewQuery(
	app *tview.Application,
	bqClient *bigquery.Client,
	checkpoint *checkpoint.Checkpoint,
	history history.Storage,
) *Query {
	q := &Query{
		Grid: tview.NewGrid(),

		app: app,

		bqClient:         bqClient,
		defaultRenderer:  &renderer.TableRenderer{},
		markdownRenderer: &renderer.MarkdownRenderer{},
		tsvRenderer:      &renderer.TSVRenderer{},
		checkpoint:       checkpoint,
		history:          history,

		textArea:          tview.NewTextArea(),
		borderTextView:    tview.NewTextView(),
		resultTextView:    tview.NewTextView(),
		statusTextView:    tview.NewTextView(),
		ctrlXTextView:     tview.NewTextView(),
		cursorPosTextView: tview.NewTextView(),

		ctrlXMode:  false,
		lastResult: nil,
	}

	q.SetRows(0, 1, 0, 1)
	q.SetColumns(0, 8, 18)

	q.AddItem(q.textArea, 0, 0, 1, 1, 0, 0, true)
	q.AddItem(q.borderTextView, 1, 0, 1, 1, 0, 0, false)
	q.AddItem(q.resultTextView, 2, 0, 1, 1, 0, 0, false)
	q.AddItem(q.statusTextView, 3, 0, 1, 1, 0, 0, false)
	q.AddItem(q.ctrlXTextView, 3, 1, 1, 1, 0, 0, false)
	q.AddItem(q.cursorPosTextView, 3, 2, 1, 1, 0, 0, false)

	return q
}

func (q *Query) Init() error {
	q.textArea.SetTextStyle(textStyleDefault).SetWordWrap(false)

	q.borderTextView.SetText("--- result ---")

	q.resultTextView.SetTextStyle(textStyleDefault).SetWordWrap(false).SetChangedFunc(func() {
		q.app.Draw()
	})

	q.statusTextView.SetTextStyle(textStyleDefault).SetChangedFunc(func() {
		q.app.Draw()
	})

	q.ctrlXTextView.SetTextStyle(textStyleDefault.Bold(true)).SetTextAlign(tview.AlignRight).SetChangedFunc(func() {
		q.app.Draw()
	})

	q.cursorPosTextView.SetTextStyle(textStyleDefault.Bold(true)).SetTextAlign(tview.AlignRight).SetChangedFunc(func() {
		q.app.Draw()
	})

	q.textArea.SetMovedFunc(func() {
		row, col, _, _ := q.textArea.GetCursor()
		// row and col starts from 0
		q.cursorPosTextView.SetText(fmt.Sprintf("(Ln %d, Col %d)", row+1, col+1))
	})

	query, err := q.checkpoint.Load()
	if err != nil {
		// ignore error
		query = ""
	}
	q.textArea.SetText(query, false)

	q.bindKeys()

	return nil
}

func (q *Query) Close() error {
	if err := q.checkpoint.Save(q.textArea.GetText(), time.Now()); err != nil {
		return fmt.Errorf("save checkpoint: %w", err)
	}

	return nil
}

func (q *Query) bindKeys() {
	// TODO: Is there a better way to propagate *global* context?
	ctx := context.Background()

	q.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if q.ctrlXMode {
			q.ctrlXMode = false
			q.ctrlXTextView.SetText("")

			switch event.Key() {
			case tcell.KeyCtrlC:
				q.app.Stop()

			case tcell.KeyEnter:
				query := q.textArea.GetText()
				q.runQuery(ctx, query, false)

			case tcell.KeyRune:
				switch event.Rune() {
				case 'c':
					q.copyResultToClipboard()

				case 'd':
					query := q.textArea.GetText()
					q.runQuery(ctx, query, true)

				case 'm':
					q.copyResultToClipboardAsMarkdown()

				case 't':
					q.copyResultToClipboardAsTSV()
				}
			default:
				// do nothing
			}

			return nil
		} else {
			switch event.Key() {
			case tcell.KeyCtrlUnderscore:
				return tcell.NewEventKey(tcell.KeyCtrlZ, 0, tcell.ModNone)

			case tcell.KeyCtrlB:
				return tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)

			case tcell.KeyCtrlC:
				q.copyQueryToClipboard()

				return nil

			case tcell.KeyCtrlF:
				return tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)

			case tcell.KeyCtrlN:
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)

			case tcell.KeyCtrlP:
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)

			case tcell.KeyCtrlX:
				q.ctrlXMode = true
				q.ctrlXTextView.SetText("Ctrl-X")

				return nil
			}
		}

		return event
	})
}

func (q *Query) runQuery(ctx context.Context, query string, dryRun bool) {
	msgPrefix := ""
	if dryRun {
		msgPrefix = "[dry-run] "
	}

	q.statusTextView.
		SetText(fmt.Sprintf("%srunning query...", msgPrefix)).
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
				q.statusTextView.
					SetText(fmt.Sprintf("%srunning query (%ds)...", msgPrefix, elapsedSecond)).
					SetTextStyle(textStyleDefault)
				elapsedSecond += 1
			}
		}
	}()

	go func() {
		start := time.Now()

		result := ""

		if dryRun {
			r, err := q.bqClient.DryRunQuery(ctx, query)
			if err != nil {
				done <- true
				q.resultTextView.SetText(err.Error())
				q.statusTextView.
					SetText(fmt.Sprintf("[ERROR] %scannot run query", msgPrefix)).
					SetTextStyle(textStyleError)

				return
			}

			result = fmt.Sprintf("This query will process %s of data.", humanize.Bytes(uint64(r.TotalBytesProcessed)))

			if err := q.history.Append(r); err != nil {
				done <- true
				q.statusTextView.
					SetText(err.Error()).
					SetTextStyle(textStyleError)

				return
			}

			done <- true

			q.statusTextView.
				SetText(fmt.Sprintf("[SUCCESS] %stook %.2f seconds", msgPrefix, time.Since(start).Seconds())).
				SetTextStyle(textStyleSuceess)
		} else {
			r, err := q.bqClient.RunQuery(ctx, query)
			if err != nil {
				done <- true
				q.resultTextView.SetText(err.Error())
				q.statusTextView.
					SetText(fmt.Sprintf("[ERROR] %scannot run query", msgPrefix)).
					SetTextStyle(textStyleError)

				return
			}

			result, err = q.defaultRenderer.Render(r)
			if err != nil {
				done <- true
				q.statusTextView.
					SetText(fmt.Sprintf("[ERROR] %scannot render result", msgPrefix)).
					SetTextStyle(textStyleError)

				return
			}

			if err := q.history.Append(r); err != nil {
				done <- true
				q.statusTextView.
					SetText(err.Error()).
					SetTextStyle(textStyleError)

				return
			}

			done <- true

			q.lastResult = r

			q.statusTextView.
				SetText(
					fmt.Sprintf(
						"[SUCCESS] %s%d row(s), took %.2f seconds, processed %s of data",
						msgPrefix,
						len(r.Rows),
						time.Since(start).Seconds(),
						humanize.Bytes(uint64(r.TotalBytesProcessed)),
					),
				).
				SetTextStyle(textStyleSuceess)
		}

		q.resultTextView.SetText(result)
		q.resultTextView.ScrollToBeginning()
	}()
}

func (q *Query) copyQueryToClipboard() {
	if err := clipboard.WriteAll(q.textArea.GetText()); err != nil {
		q.statusTextView.
			SetText(fmt.Sprintf("cannot copy query to clipboard: %s", err)).
			SetTextStyle(textStyleError)

		return
	}

	q.statusTextView.SetText("copied query to clipboard").SetTextStyle(textStyleSuceess)
}

func (q *Query) copyResultToClipboard() {
	if err := clipboard.WriteAll(q.resultTextView.GetText(true)); err != nil {
		q.statusTextView.
			SetText(fmt.Sprintf("cannot copy result to clipboard: %s", err)).
			SetTextStyle(textStyleError)

		return
	}

	q.statusTextView.SetText("copied result to clipboard").SetTextStyle(textStyleSuceess)
}

func (q *Query) copyResultToClipboardAsMarkdown() {
	if q.lastResult == nil {
		q.statusTextView.SetText("nothing to copy").SetTextStyle(textStyleError)
		return
	}

	t, err := q.markdownRenderer.Render(q.lastResult)
	if err != nil {
		q.statusTextView.
			SetText(fmt.Sprintf("cannot render result as Markdown table: %s", err)).
			SetTextStyle(textStyleError)

		return
	}

	if err := clipboard.WriteAll(t); err != nil {
		q.statusTextView.
			SetText(fmt.Sprintf("cannot copy result to clipboard as Markdown table: %s", err)).
			SetTextStyle(textStyleError)

		return
	}

	q.statusTextView.SetText("copied result to clipboard as Markdown table").SetTextStyle(textStyleSuceess)
}

func (q *Query) copyResultToClipboardAsTSV() {
	if q.lastResult == nil {
		q.statusTextView.SetText("nothing to copy").SetTextStyle(textStyleError)
		return
	}

	t, err := q.tsvRenderer.Render(q.lastResult)
	if err != nil {
		q.statusTextView.
			SetText(fmt.Sprintf("cannot render result as TSV: %s", err)).
			SetTextStyle(textStyleError)

		return
	}

	if err := clipboard.WriteAll(t); err != nil {
		q.statusTextView.
			SetText(fmt.Sprintf("cannot copy result to clipboard as TSV: %s", err)).
			SetTextStyle(textStyleError)

		return
	}

	q.statusTextView.SetText("copied result to clipboard as TSV").SetTextStyle(textStyleSuceess)
}
