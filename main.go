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

	"github.com/dtan4/bqc/internal/bigquery"
	"github.com/dtan4/bqc/internal/renderer"
	"github.com/dtan4/bqc/internal/screen"
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

	scr := screen.New(client, rdr)

	if err := scr.Run(ctx); err != nil {
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
