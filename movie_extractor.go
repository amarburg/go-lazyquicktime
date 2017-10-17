package lazyquicktime

import (
	"image"
)

// MovieExtractor is the abstract interface to a quicktime movie.
type MovieExtractor interface {
	NumFrames() int
	Duration() float32
	ExtractFrame(frame int) (image.Image, error)
}
