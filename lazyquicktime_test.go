package lazyquicktime

import "testing"
import "fmt"
//import "io"
import "os"

import "github.com/amarburg/go-lazyfs"

import "image/png"

//import "net/url"
//var TestUrlRoot = "https://amarburg.github.io/go-lazyfs-testfiles/"
//var TestUrlRoot = "http://localhost:8080/files/"
//var TestUrl,_ = url.Parse( TestUrlRoot + TestMovPath )
//var TestMovPath = "CamHD_Vent_Short.mov"


// For local testing
import "github.com/amarburg/go-lazyfs-testfiles"
var TestMovPath = lazyfs_testfiles.TestMovPath


var SparseHttpStoreRoot = "cache/httpsparse/"

func TestConvert( t *testing.T ) {

  //source,err := lazyfs.OpenHttpSource( *TestUrl )
  source,err := lazyfs.OpenLocalFileSource( "../go-lazyfs-testfiles/", TestMovPath )
  if err != nil {
    panic("Couldn't open HttpFSSource")
  }

  store,err := lazyfs.OpenSparseFileStore( source, SparseHttpStoreRoot )
  if store == nil {
    panic("Couldn't open SparesFileFSStore")
  }

  mov := LoadMovMetadata( store )

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
  img,_ := mov.ExtractFrame( frame )

  if err != nil { panic(fmt.Sprintf("Error decoding frame: %s", err.Error()))}

  img_filename := fmt.Sprintf("frame%06d.png", frame)
  img_file,err := os.Create(img_filename)
  if err != nil { panic(fmt.Sprintf("Error creating png %s: %s", img_filename, err.Error()))}

  err = png.Encode( img_file, img )
  if err != nil { panic(fmt.Sprintf("Error writing png %s: %s", img_filename, err.Error()))}

}
