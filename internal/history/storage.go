package history

import (
	"encoding/gob"
	"time"

	bigqueryapi "cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"

	"github.com/dtan4/bqc/internal/bigquery"
)

type Storage interface {
	Close() error
	Append(result *bigquery.Result) error
	List() ([]*bigquery.Result, error)
}

func init() {
	gob.Register(map[string]bigqueryapi.Value{})
	gob.Register([]bigqueryapi.Value{})
	gob.Register(time.Time{})
	gob.Register(civil.Date{})
	gob.Register(civil.DateTime{})
}
