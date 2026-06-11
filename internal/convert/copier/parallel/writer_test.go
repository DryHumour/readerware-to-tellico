package parallel

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/DryHumour/readerware-to-tellico/internal/tellico/tcfile"
)

func TestWriterRun(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("happy path", func(t *testing.T) {
		var buf bytes.Buffer
		tcf := tcfile.New(&buf)

		resultC := make(chan result, 10)
		writerC := make(chan *payload)

		w := newWriter(logger, tcf, resultC, writerC)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		// Create a dummy image and payload
		img, _ := tcfile.NewImage("test.jpg", dummyFileInfo{name: "test.jpg", size: 4}, "comment")
		p := acquirePayload()
		p.ID = "test.jpg"
		p.Image = img
		p.Data = []byte("data")

		go w.Run(ctx)
		writerC <- p
		close(writerC)

		// Check results
		res := <-resultC
		if res.Report.Message != "writing id=\"test.jpg\"" {
			t.Errorf("unexpected progress message: %q", res.Report.Message)
		}

		err := tcf.Close()
		if err != nil {
			t.Fatal(err)
		}

		if buf.Len() == 0 {
			t.Error("expected data to be written to buffer")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		var buf bytes.Buffer
		tcf := tcfile.New(&buf)
		resultC := make(chan result)
		writerC := make(chan *payload)

		w := newWriter(logger, tcf, resultC, writerC)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		w.Run(ctx) // should return immediately
	})
}

type dummyFileInfo struct {
	name string
	size int64
}

func (d dummyFileInfo) Name() string       { return d.name }
func (d dummyFileInfo) Size() int64        { return d.size }
func (d dummyFileInfo) Mode() os.FileMode  { return 0 }
func (d dummyFileInfo) ModTime() time.Time { return time.Now() }
func (d dummyFileInfo) IsDir() bool        { return false }
func (d dummyFileInfo) Sys() interface{}   { return nil }
