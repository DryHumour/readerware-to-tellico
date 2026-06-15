package parallel

import (
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/convert/copier"
)

func TestPayloadPool(t *testing.T) {
	t.Parallel()

	// Test acquirePayload
	p := acquirePayload()
	if p == nil {
		t.Fatal("acquirePayload returned nil")
	}
	if cap(p.Data) != copier.MaxReaderwareImageSize {
		t.Errorf("expected buffer capacity %d, got %d", copier.MaxReaderwareImageSize, cap(p.Data))
	}

	// Test releasePayload with normal buffer
	p.ID = "test-id"
	p.Data = p.Data[:100] // simulate partial use
	releasePayload(p)

	// Acquire again and check if it's the same or reset
	p2 := acquirePayload()
	if len(p2.Data) != copier.MaxReaderwareImageSize {
		t.Errorf("expected buffer length %d after release, got %d", copier.MaxReaderwareImageSize, len(p2.Data))
	}

	// Test releasePayload with nil
	releasePayload(nil) // should not panic

	// Test releasePayload with oversized buffer
	oversized := &payload{
		Data: make([]byte, copier.MaxReaderwareImageSize+1),
	}
	releasePayload(oversized)
	// We can't easily check if it was dropped from the pool without internal access,
	// but we've verified it doesn't panic.
}
