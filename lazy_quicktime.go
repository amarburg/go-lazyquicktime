package lazyquicktime

import "fmt"
import "image"
import "errors"

import "time"
import "log"

import "github.com/amarburg/go-lazyfs"
import "github.com/amarburg/go-quicktime"
import "github.com/amarburg/go-prores-ffmpeg"

// Version string for the package
const Version = "v0.1.0"

// LazyQuicktime stores the metadata from a Quicktime Movie
// extracted via a lazyFS.
// Stores a copy of the lazyFS to allow lazy-loading
type LazyQuicktime struct {
	file lazyfs.FileSource
	Tree quicktime.AtomArray
	Trak quicktime.TRAKAtom
	Stbl *quicktime.STBLAtom
	Mvhd quicktime.MVHDAtom

	FileSize int64
}

// LoadMovMetadata creates a LazyQuicktime by querying a lazyfs.FileSource.
func LoadMovMetadata(file lazyfs.FileSource) (*LazyQuicktime, error) {

	mov := &LazyQuicktime{file: file}

	sz, err := file.FileSize()
	if sz < 0 || err != nil {
		return mov, fmt.Errorf("Unable to retrieve file size.")
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
		return mov, errors.New("Can't find MOOV atom")
	}

	mvhd := moov.FindAtom("mvhd")
	if mvhd == nil {
		return mov, errors.New("Couldn't find MVHD in the moov atom")
	}
	mov.Mvhd, _ = quicktime.ParseMVHD(mvhd)

	tracks := moov.FindAtoms("trak")
	if tracks == nil || len(tracks) == 0 {
		return mov, errors.New("Couldn't find any TRAKs in the MOOV")
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
		return mov, errors.New("Couldn't identify the Video track")
	}

	mov.Trak, err = quicktime.ParseTRAK(track)
	if err != nil {
		return mov, fmt.Errorf("Unable to parse TRAK atom: %s", err.Error())
	}

	mov.Stbl = &mov.Trak.Mdia.Minf.Stbl // Just an alias

	return mov, nil
}

// NumFrames reports the number of frames in the LazyQuicktime
func (mov *LazyQuicktime) NumFrames() int {
	return mov.Stbl.NumFrames()
}

// Duration reports the length of the LazyQuicktime in seconds.
func (mov *LazyQuicktime) Duration() float32 {
	return mov.Mvhd.Duration()
}

// ExtractFrame extracts an individual frame from a ProRes file as an Image
func (mov *LazyQuicktime) ExtractFrame(frame int) (image.Image, error) {
	return mov.ExtractNRGBA(frame)
}

// ExtractNRGBA extracts an individual frame from a ProRes file specifically
// as an image.NRGBA
func (mov *LazyQuicktime) ExtractNRGBA(frame int) (*image.NRGBA, error) {

	frameOffset, frameSize, _ := mov.Stbl.SampleOffsetSize(frame)

	//fmt.Printf("Extracting frame %d at offset %d size %d\n", frame, frame_offset, frame_size)

	buf := make([]byte, frameSize)

	if buf == nil {
		return nil, fmt.Errorf("Couldn't make buffer of size %d", frameSize)
	}

	startRead := time.Now()
	n, _ := mov.file.ReadAt(buf, frameOffset)
	log.Printf("HTTP read took %s", time.Since(startRead))

	if n != frameSize {
		return nil, fmt.Errorf("Tried to read %d bytes but got %d instead", frameSize, n)
	}

	width, height := int(mov.Trak.Tkhd.Width), int(mov.Trak.Tkhd.Height)

	//startDecode := time.Now()
	img, err := prores.DecodeProRes(buf, width, height)
	//log.Printf("Prores decode took %s", time.Since(startDecode))

	return img, err

}
