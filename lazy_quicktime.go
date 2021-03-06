package lazyquicktime

import "fmt"
import "image"
import "errors"

import "time"

import "github.com/amarburg/go-lazyfs"
import "github.com/amarburg/go-quicktime"
import "github.com/amarburg/go-prores-ffmpeg"

// Version string for the package
const Version = "v0.1.0"

// LazyQuicktime stores the metadata from a Quicktime Movie
// extracted via a lazyFS.
// Stores a copy of the lazyFS to allow lazy-loading
type LazyQuicktime struct {
	Source lazyfs.FileSource
	Tree   quicktime.AtomArray
	Trak   quicktime.TRAKAtom
	Stbl   *quicktime.STBLAtom
	Mvhd   quicktime.MVHDAtom

	FileSize int64
}

// LoadMovMetadata creates a LazyQuicktime by querying a lazyfs.FileSource.
func LoadMovMetadata(file lazyfs.FileSource) (*LazyQuicktime, error) {

	mov := &LazyQuicktime{Source: file}

	sz, err := file.FileSize()
	if sz < 0 || err != nil {
		return mov, fmt.Errorf("unable to retrieve file size")
	}

	mov.FileSize = sz

	setEagerload := func(conf *quicktime.BuildTreeConfig) {
		conf.EagerloadTypes = []string{"moov"}
	}

	//fmt.Println("Reading Mov of size ", mov.FileSize)
	tree, err := quicktime.BuildTree(file, uint64(mov.FileSize), setEagerload)

	if err != nil {
		return mov, err
	}
	mov.Tree = tree

	moov := mov.Tree.FindAtom("moov")
	if moov == nil {
		return mov, errors.New("can't find MOOV atom")
	}

	mvhd := moov.FindAtom("mvhd")
	if mvhd == nil {
		return mov, errors.New("couldn't find MVHD in the moov atom")
	}
	mov.Mvhd, _ = quicktime.ParseMVHD(mvhd)

	tracks := moov.FindAtoms("trak")
	if tracks == nil || len(tracks) == 0 {
		return mov, errors.New("couldn't find any TRAKs in the MOOV")
	}

	var track *quicktime.Atom
	for i, t := range tracks {
		mdia := t.FindAtom("mdia")
		if mdia == nil {
			fmt.Println("No mdia track", i)
			continue
		}

		minf := mdia.FindAtom("minf")
		if minf == nil {
			fmt.Println("No minf track", i)
			continue
		}

		if minf.FindAtom("vmhd") != nil {
			track = t
			break
		}
	}

	if track == nil {
		return mov, errors.New("couldn't identify the Video track")
	}

	mov.Trak, err = quicktime.ParseTRAK(track)
	if err != nil {
		return mov, fmt.Errorf("unable to parse TRAK atom: %s", err.Error())
	}

	mov.Stbl = &mov.Trak.Mdia.Minf.Stbl // Just an alias

	return mov, nil
}

// NumFrames reports the number of frames in the LazyQuicktime
func (mov *LazyQuicktime) NumFrames() uint64 {
	return mov.Stbl.NumFrames()
}

// Duration reports the length of the LazyQuicktime in seconds.
func (mov *LazyQuicktime) Duration() time.Duration {
	return mov.Mvhd.Duration()
}

type LQTPerformance struct {
		Read, Decode   time.Duration
}

// ExtractFrame extracts an individual frame from a ProRes file as an Image
func (mov *LazyQuicktime) ExtractFrame(frame uint64) (image.Image, error) {
	return mov.ExtractNRGBA(frame)
}

// ExtractFrame extracts an individual frame from a ProRes file as an Image
func (mov *LazyQuicktime) ExtractFramePerf(frame uint64) (image.Image, LQTPerformance, error) {
	return mov.ExtractNRGBAPerf(frame)
}

// ExtractNRGBA extracts an individual frame from a ProRes file as an image.NRGBA
// in its native height
func (mov *LazyQuicktime) ExtractNRGBA(frame uint64) (*image.NRGBA, error) {
	frameOffset, frameSize, _ := mov.Stbl.SampleOffsetSize(int(frame))

	buf := make([]byte, frameSize)

	if buf == nil {
		return nil, fmt.Errorf("couldn't make buffer of size %d", frameSize)
	}

	n, _ := mov.Source.ReadAt(buf, frameOffset)

	if n != frameSize {
		return nil, fmt.Errorf("tried to read %d bytes but got %d instead", frameSize, n)
	}

	img, err := prores.DecodeProRes(buf, int(mov.Trak.Tkhd.Width), int(mov.Trak.Tkhd.Height))

	return img, err
}

// ExtractNRGBAPerf extracts an individual frame from a ProRes file as an image.NRGBA
// and also returns performance information in an LQTPerformance structure
func (mov *LazyQuicktime) ExtractNRGBAPerf(frame uint64) (*image.NRGBA, LQTPerformance, error) {

var perf LQTPerformance

	frameOffset, frameSize, _ := mov.Stbl.SampleOffsetSize(int(frame))

	buf := make([]byte, frameSize)

	if buf == nil {
		return nil, perf, fmt.Errorf("couldn't make buffer of size %d", frameSize)
	}

	startRead := time.Now()
	n, _ := mov.Source.ReadAt(buf, frameOffset)
	perf.Read = time.Since(startRead);

	if n != frameSize {
		return nil, perf, fmt.Errorf("tried to read %d bytes but got %d instead", frameSize, n)
	}

	width, height := int(mov.Trak.Tkhd.Width), int(mov.Trak.Tkhd.Height)

	startDecode := time.Now()
	img, err := prores.DecodeProRes(buf, width, height)
	perf.Decode = time.Since(startDecode)

	return img, perf, err

}
