package renderer

import (
	"strings"
	"testing"
	"time"

	bigqueryapi "cloud.google.com/go/bigquery"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/dtan4/bqc/internal/bigquery"
)

func TestTableRender(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		result *bigquery.Result
		want   string
	}{
		"success": {
			result: &bigquery.Result{
				Keys: []string{
					"foo",
					"bar",
					"baz",
				},
				Rows: []map[string]bigqueryapi.Value{
					{
						"foo": "foovalue",
						"bar": 1,
						"baz": time.Date(2023, 5, 3, 12, 34, 56, 0, time.UTC),
					},
				},
			},
			want: `+----------+-----+-------------------------------+
|   foo    | bar |              baz              |
+----------+-----+-------------------------------+
| foovalue |   1 | 2023-05-03 12:34:56 +0000 UTC |
+----------+-----+-------------------------------+
`,
		},
		"long text won't be wrapped": {
			result: &bigquery.Result{
				Keys: []string{
					"foo",
					"bar",
					"baz",
					"qux",
					"quux",
				},
				Rows: []map[string]bigqueryapi.Value{
					{
						"foo":  "foovalue",
						"bar":  1,
						"baz":  time.Date(2023, 5, 3, 12, 34, 56, 0, time.UTC),
						"qux":  "qux",
						"quux": "quuuuuu uuuuuuuuuuuuuu uuuuuuuuux",
					},
				},
			},
			want: `+----------+-----+-------------------------------+-----+-----------------------------------+
|   foo    | bar |              baz              | qux |               quux                |
+----------+-----+-------------------------------+-----+-----------------------------------+
| foovalue |   1 | 2023-05-03 12:34:56 +0000 UTC | qux | quuuuuu uuuuuuuuuuuuuu uuuuuuuuux |
+----------+-----+-------------------------------+-----+-----------------------------------+
`,
		},
	}

	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rdr := &TableRenderer{}

			got, err := rdr.Render(tc.result)
			if err != nil {
				t.Errorf("want no error, got: %s", err)
			}

			if diff := cmp.Diff(tc.want, got, cmpopts.AcyclicTransformer("SplitLines", func(s string) []string {
				return strings.Split(s, "\n")
			})); diff != "" {
				t.Errorf("Render() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMarkdownRender(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		result *bigquery.Result
		want   string
	}{
		"success": {
			result: &bigquery.Result{
				Keys: []string{
					"foo",
					"bar",
					"baz",
				},
				Rows: []map[string]bigqueryapi.Value{
					{
						"foo": "foovalue",
						"bar": 1,
						"baz": time.Date(2023, 5, 3, 12, 34, 56, 0, time.UTC),
					},
				},
			},
			want: `|   foo    | bar |              baz              |
|----------|-----|-------------------------------|
| foovalue |   1 | 2023-05-03 12:34:56 +0000 UTC |
`,
		},
		"long text won't be wrapped": {
			result: &bigquery.Result{
				Keys: []string{
					"foo",
					"bar",
					"baz",
					"qux",
					"quux",
				},
				Rows: []map[string]bigqueryapi.Value{
					{
						"foo":  "foovalue",
						"bar":  1,
						"baz":  time.Date(2023, 5, 3, 12, 34, 56, 0, time.UTC),
						"qux":  "qux",
						"quux": "quuuuuu uuuuuuuuuuuuuu uuuuuuuuux",
					},
				},
			},
			want: `|   foo    | bar |              baz              | qux |               quux                |
|----------|-----|-------------------------------|-----|-----------------------------------|
| foovalue |   1 | 2023-05-03 12:34:56 +0000 UTC | qux | quuuuuu uuuuuuuuuuuuuu uuuuuuuuux |
`,
		},
	}

	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rdr := &MarkdownRenderer{}

			got, err := rdr.Render(tc.result)
			if err != nil {
				t.Errorf("want no error, got: %s", err)
			}

			if diff := cmp.Diff(tc.want, got, cmpopts.AcyclicTransformer("SplitLines", func(s string) []string {
				return strings.Split(s, "\n")
			})); diff != "" {
				t.Errorf("Render() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
