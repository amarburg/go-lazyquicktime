package main

import "fmt"
import "io"
import "os"

import "github.com/amarburg/go-lazyfs"
import "github.com/amarburg/go-quicktime"
import "github.com/amarburg/go-prores-ffmpeg"

import "image/png"

var TestUrlRoot = "https://amarburg.github.io/go-lazyfs-testfiles/"
//var TestUrlRoot = "http://localhost:8080/files/"
var TestMovPath = "CamHD_Vent_Short.mov"

var SparseHttpStoreRoot = "cache/httpsparse/"

func main() {

  source,err := lazyfs.OpenHttpFSSource(TestUrlRoot)
  if err != nil {
    panic("Couldn't open HttpFSSource")
  }

  store,err := lazyfs.OpenSparseFileFSStore( SparseHttpStoreRoot )
  if store == nil {
    panic("Couldn't open SparesFileFSStore")
  }

  source.SetBackingStore( store )

  file,err := source.Open( TestMovPath )
  if err != nil {
    panic("Couldn't open AlphabetPath")
  }

  sz,_ := file.FileSize()

  set_eagerload := func ( conf *quicktime.BuildTreeConfig ) {
      conf.EagerloadTypes = []string{"moov"}
  }

  tree,err := quicktime.BuildTree( file, sz, set_eagerload )
  if err != nil {
    panic("Couldn't build Tree")
  }

  quicktime.DumpTree( tree )

  moov := tree.FindAtom("moov")
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

  trak,err := quicktime.ParseTRAK( track )
  if err != nil { panic(fmt.Sprintf("Unable to parse TRAK atom: %s", err.Error()))}
  stbl := &trak.Mdia.Minf.Stbl          // Just an alias

  //fmt.Println("Found track with video information")

  // Find movie length
  num_frames := stbl.NumFrames()

  fmt.Println("Movie has",num_frames,"frames")

  // fmt.Println("Chunk table:")
  // for idx,offset := range stbl.Stco.ChunkOffsets {
  //   fmt.Printf("   %d %20d\n",idx+1, offset )
  // }

  //fmt.Println(stbl)

  // for sample := 1; sample <= num_frames; sample++ {
  //     chunk,chunk_start,relasample := stbl.Stsc.SampleChunk( sample )
  //     fmt.Println("Sample", sample,"is in chunk",chunk,"the",relasample,"'th sample; the chunk starts at sample",chunk_start)
  //
  //     offset,_ := stbl.SampleOffset( sample )
  //     fmt.Println("Sample at byte",offset,"in file")
  //
  //
  // }

  // Try extracting a frame
  frame := 2
  LoadFrame( frame, trak, file )

}


func LoadFrame( frame int, trak quicktime.TRAKAtom, file io.ReaderAt ) {

  frame_offset,frame_size,_ := trak.Mdia.Minf.Stbl.SampleOffsetSize( frame )

  fmt.Printf("Extracting frame %d at offset %d size %d\n", frame, frame_offset, frame_size)

  buf := make([]byte, frame_size)
  n,_ := file.ReadAt( buf, frame_offset )

  if n != frame_size { panic(fmt.Sprintf("Tried to read %d bytes but got %d instead",frame_size,n))}

  width, height := int(trak.Tkhd.Width), int(trak.Tkhd.Height)
  fmt.Printf("Image is %d x %d", width, height)

  img,err := prores.DecodeProRes( buf, width, height )

  if err != nil { panic(fmt.Sprintf("Error decoding frame: %s", err.Error()))}

  img_filename := fmt.Sprintf("frame%06d.png", frame)
  img_file,err := os.Create(img_filename)
  if err != nil { panic(fmt.Sprintf("Error creating png %s: %s", img_filename, err.Error()))}

  err = png.Encode( img_file, img )
  if err != nil { panic(fmt.Sprintf("Error writing png %s: %s", img_filename, err.Error()))}

}
