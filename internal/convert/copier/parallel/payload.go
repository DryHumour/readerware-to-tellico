package parallel

import (
	"sync"

	"github.com/DryHumour/readerware-to-tellico/internal/convert/copier"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/tcfile"
)

var (
	// payloadPool is a sync.Pool for reusing payload objects to reduce GC pressure.
	// Oversized buffers are dropped in releasePayload instead of being returned.
	payloadPool = sync.Pool{
		New: func() interface{} {
			return &payload{Data: make([]byte, copier.MaxReaderwareImageSize)}
		},
	}
)

// payload represents an image and its data being transferred from reader to writer.
// The Data field contains the raw image bytes, which may be larger than copierpkg.MaxReaderwareImageSize
// for oversized files (which are not returned to the pool).
type payload struct {
	ID    string
	Image tcfile.Image
	Data  []byte
}

// acquirePayload obtains a payload from the pool with a pre-allocated buffer of copierpkg.MaxReaderwareImageSize.
// The buffer is reset to full capacity before use.
func acquirePayload() *payload {
	return payloadPool.Get().(*payload)
}

// releasePayload returns a payload to the pool for reuse.
// Oversized buffers (those with capacity != copierpkg.MaxReaderwareImageSize) are dropped instead.
// The payload buffer is reset to full capacity before returning.
func releasePayload(payload *payload) {
	if payload == nil {
		return
	}
	if cap(payload.Data) != copier.MaxReaderwareImageSize {
		// drop oversized buffers
		return
	}
	payload.Data = payload.Data[:cap(payload.Data)]
	payloadPool.Put(payload)
}
