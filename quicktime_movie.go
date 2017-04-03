package lazyquicktime

import "fmt"
import "image"
import "errors"

import "github.com/amarburg/go-lazyfs"
import "github.com/amarburg/go-quicktime"
import "github.com/amarburg/go-prores-ffmpeg"

type LazyQuicktime struct {
	file lazyfs.FileSource
	Tree quicktime.AtomArray
	Trak quicktime.TRAKAtom
	Stbl *quicktime.STBLAtom
	Mvhd quicktime.MVHDAtom
}

func LoadMovMetadata(file lazyfs.FileSource) (*LazyQuicktime, error) {

	sz, _ := file.FileSize()

	set_eagerload := func(conf *quicktime.BuildTreeConfig) {
		conf.EagerloadTypes = []string{"moov"}
	}

	mov := &LazyQuicktime{file: file}

	tree, err := quicktime.BuildTree(file, sz, set_eagerload)

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

	var track *quicktime.Atom = nil
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
		return mov, errors.New(fmt.Sprintf("Unable to parse TRAK atom: %s", err.Error()))
	}

	mov.Stbl = &mov.Trak.Mdia.Minf.Stbl // Just an alias

	return mov, nil
}

func (mov *LazyQuicktime) NumFrames() int {
	return mov.Stbl.NumFrames()
}

func (mov *LazyQuicktime) Duration() float32 {
	return mov.Mvhd.Duration()
}

func (mov *LazyQuicktime) ExtractFrame(frame int) (image.Image, error) {

	frame_offset, frame_size, _ := mov.Stbl.SampleOffsetSize(frame)

	fmt.Printf("Extracting frame %d at offset %d size %d\n", frame, frame_offset, frame_size)

	buf := make([]byte, frame_size)

	if buf == nil {
		return nil, fmt.Errorf("Couldn't make buffer of size %d", frame_size)
	}

	n, _ := mov.file.ReadAt(buf, frame_offset)

	if n != frame_size {
		return nil, fmt.Errorf("Tried to read %d bytes but got %d instead", frame_size, n)
	}

	width, height := int(mov.Trak.Tkhd.Width), int(mov.Trak.Tkhd.Height)

	img, err := prores.DecodeProRes(buf, width, height)

	return img, err

}
