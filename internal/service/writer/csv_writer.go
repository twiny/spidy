package writer

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/twiny/spidy/v2/internal/pkg/spider/v1"
)

type CSVWriter struct {
	l *sync.Mutex
	f *os.File
	w *csv.Writer
}

func NewCSVWriter(dir string) (*CSVWriter, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	name := time.Now().Format("2006-01-02")
	fp := filepath.Join(dir, name+"_domains.csv")

	// open or create log
	f, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &CSVWriter{
		l: &sync.Mutex{},
		f: f,
		w: csv.NewWriter(f),
	}, nil
}

func (c *CSVWriter) Write(d *spider.Domain) error {
	c.l.Lock()
	defer func() {
		c.l.Unlock()
		c.w.Flush()
	}()

	return c.w.Write([]string{d.Name + "." + d.TLD, d.Status})
}

func (c *CSVWriter) Close() error {
	return c.f.Close()
}
