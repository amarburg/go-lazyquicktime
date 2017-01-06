package lazyquicktime

import "fmt"
import "image"

import "github.com/amarburg/go-lazyfs"
import "github.com/amarburg/go-quicktime"
import "github.com/amarburg/go-prores-ffmpeg"


type LazyQuicktime struct {
  file  lazyfs.File
  Tree  quicktime.AtomArray
  Trak  quicktime.TRAKAtom
  Stbl  *quicktime.STBLAtom
}


func LoadMovMetadata( file lazyfs.File ) (*LazyQuicktime ) {

  sz,_ := file.FileSize()

  set_eagerload := func ( conf *quicktime.BuildTreeConfig ) {
    conf.EagerloadTypes = []string{"moov"}
  }

  mov := LazyQuicktime{ file: file }
  tree,err := quicktime.BuildTree( file, sz, set_eagerload )
  if err != nil {
    panic("Couldn't build Tree")
  }
  mov.Tree = tree


  //quicktime.DumpTree( mov.Tree )

  moov := mov.Tree.FindAtom("moov")
  moov.ReadData( file )
  if moov == nil { panic("Can't find MOOV atom")}

  tracks := moov.FindAtoms("trak")
  if tracks == nil || len(tracks) == 0 { panic("Couldn't find any TRAKs in the MOOV")}
  //fmt.Println("Found",len(tracks),"TRAK atoms")

  var track *quicktime.Atom = nil
  for i,t := range tracks {
    mdia := t.FindAtom("mdia")
    if mdia == nil {
      fmt.Println("No mdia track",i)
      continue
    }

    minf := mdia.FindAtom("minf")
    if minf == nil {
      fmt.Println("No minf track",i)
      continue
    }

    if minf.FindAtom("vmhd") != nil {
      track = t
      break
    }
  }

  if track == nil { panic("Couldn't identify the Video track")}

  mov.Trak,err = quicktime.ParseTRAK( track )
  if err != nil { panic(fmt.Sprintf("Unable to parse TRAK atom: %s", err.Error()))}

  mov.Stbl = &mov.Trak.Mdia.Minf.Stbl          // Just an alias

  //num_frames := mov.Stbl.NumFrames()
  //fmt.Println("Movie has",num_frames,"frames")

  return &mov
}


func (mov *LazyQuicktime) ExtractFrame( frame int ) (image.Image,error) {

  frame_offset,frame_size,_ := mov.Stbl.SampleOffsetSize( frame )

  fmt.Printf("Extracting frame %d at offset %d size %d\n", frame, frame_offset, frame_size)

  buf := make([]byte, frame_size)
  n,_ := mov.file.ReadAt( buf, frame_offset )

  if n != frame_size { panic(fmt.Sprintf("Tried to read %d bytes but got %d instead",frame_size,n))}

  width, height := int(mov.Trak.Tkhd.Width), int(mov.Trak.Tkhd.Height)
  fmt.Printf("Image is %d x %d", width, height)

  img,err := prores.DecodeProRes( buf, width, height )

  return img,err

}
