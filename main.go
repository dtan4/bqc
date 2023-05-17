package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/dtan4/bqc/internal/bigquery"
	"github.com/dtan4/bqc/internal/renderer"
)

const (
	bigqueryConfigFilename = ".bigqueryrc"
)

var (
	projectIDConfigRe = regexp.MustCompile(`^project_id = ([a-z0-9-]+)$`)
)

func main() {
	if err := realMain(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func realMain(args []string) error {
	var projectID string

	if len(args) == 0 {
		projectID = loadProjectIDFromConfig()
	} else {
		projectID = args[0]
	}

	if projectID == "" {
		return errors.New("project ID must be provided")
	}

	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("create BigQuery client: %w", err)
	}
	defer client.Close()

	rdr := &renderer.TableRenderer{}

	app := tview.NewApplication()

	textArea := tview.NewTextArea().SetTextStyle(tcell.StyleDefault)

	borderTextView := tview.NewTextView().SetText("--- result ---")
	resultTextView := tview.NewTextView().SetTextStyle(tcell.StyleDefault).SetChangedFunc(func() {
		app.Draw()
	})
	statusTextView := tview.NewTextView().SetText("").SetChangedFunc(func() {
		app.Draw()
	})

	mainView := tview.NewGrid().
		SetRows(0, 1, 0, 1).
		AddItem(textArea, 0, 0, 1, 1, 0, 0, true).
		AddItem(borderTextView, 1, 0, 1, 1, 0, 0, false).
		AddItem(resultTextView, 2, 0, 1, 1, 0, 0, false).
		AddItem(statusTextView, 3, 0, 1, 1, 0, 0, false)

	pages := tview.NewPages().AddAndSwitchToPage("main", mainView, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Some environments interpret Ctrl+Enter as Ctrl+J
		if ((event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyCR) && event.Modifiers() == tcell.ModCtrl) ||
			event.Key() == tcell.KeyCtrlJ {
			q := textArea.GetText()

			statusTextView.
				SetText("running query...").
				SetTextStyle(tcell.StyleDefault)

			elapsedSecond := 1

			ticker := time.NewTicker(1 * time.Second)
			done := make(chan bool)

			go func() {
				for {
					select {
					case <-done:
						return
					case <-ticker.C:
						statusTextView.
							SetText(fmt.Sprintf("running query (%ds)...", elapsedSecond)).
							SetTextStyle(tcell.StyleDefault)
						elapsedSecond += 1
					}
				}
			}()

			go func() {
				start := time.Now()
				r, err := client.RunQuery(ctx, q)
				if err != nil {
					done <- true
					resultTextView.SetText(err.Error())
					statusTextView.
						SetText("[ERROR] cannot run query").
						SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))

					return
				}

				t, err := rdr.Render(r)
				if err != nil {
					done <- true
					statusTextView.
						SetText("[ERROR] cannot render result").
						SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))

					return
				}

				resultTextView.SetText(t)
				resultTextView.ScrollToBeginning()

				done <- true

				statusTextView.
					SetText(fmt.Sprintf("[SUCCESS] %d row(s), took %.2f seconds", len(r.Rows), time.Since(start).Seconds())).
					SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorGreenYellow))
			}()

			return nil
		}

		return event
	})

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		return fmt.Errorf("run TUI app: %w", err)
	}

	return nil
}

func loadProjectIDFromConfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	filename := filepath.Join(home, bigqueryConfigFilename)

	f, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer f.Close()

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		if matched := projectIDConfigRe.FindStringSubmatch(line); len(matched) > 1 {
			return matched[1]
		}
	}

	return ""
}
