package lazyquicktime

import (
	"image"
	"time"
)

// MovieExtractor is the abstract interface to a quicktime movie.
type MovieExtractor interface {
	NumFrames() uint64
	Duration() time.Duration
	ExtractFrame(frame uint64) (image.Image, error)
}
