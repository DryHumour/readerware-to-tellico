package copier_test

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/config"
	"github.com/DryHumour/readerware-to-tellico/internal/convert/copier/parallel"
	"github.com/DryHumour/readerware-to-tellico/internal/convert/copier/simple"
	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/tcfile"
)

func BenchmarkCopiers(b *testing.B) {
	ctx := b.Context()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	exportDir := os.Getenv("RW2TC_BENCHMARK_DATA")
	if exportDir == "" {
		b.Skip("skipping benchmark: RW2TC_BENCHMARK_DATA environment variable not set")
	}

	dirs := config.Directories{}.DefaultToExtracted(exportDir)
	index, err := images.BuildIndex(ctx, dirs)
	if err != nil {
		b.Fatalf("failed to build index: %v", err)
	}

	var allEntries []*images.ManifestEntry
	for entry := range index.All() {
		allEntries = append(allEntries, entry)
	}

	if len(allEntries) == 0 {
		b.Fatal("no entries found to benchmark")
	}

	b.Logf("Benchmarking with %d entries", len(allEntries))

	b.Run("Simple", func(b *testing.B) {
		for b.Loop() {
			tcf := tcfile.New(io.Discard)
			c := simple.New(logger, tcf)

			entriesIter := func(yield func(*images.ManifestEntry) bool) {
				for _, entry := range allEntries {
					if !yield(entry) {
						return
					}
				}
			}

			for _, err := range c.CopyAll(ctx, entriesIter) {
				if err != nil {
					b.Fatalf("Simple CopyAll failed: %v", err)
				}
			}
			tcf.Close()
		}
	})

	concurrencies := []int{4, 16, 32, 64, 256}
	for _, n := range concurrencies {
		b.Run(fmt.Sprintf("Parallel/%d", n), func(b *testing.B) {
			for b.Loop() {
				tcf := tcfile.New(io.Discard)
				c := parallel.New(logger, tcf, n)

				entriesIter := func(yield func(*images.ManifestEntry) bool) {
					for _, entry := range allEntries {
						if !yield(entry) {
							return
						}
					}
				}

				for _, err := range c.CopyAll(ctx, entriesIter) {
					if err != nil {
						b.Fatalf("Parallel/%d CopyAll failed: %v", n, err)
					}
				}
				tcf.Close()
			}
		})
	}
}
