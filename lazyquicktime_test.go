package lazyquicktime

import "testing"
import "fmt"
import "net/url"
//import "io"
import "os"

import "github.com/amarburg/go-lazyfs"

import "image/png"
import "github.com/amarburg/go-lazyfs-testfiles/http_server"


//import "net/url"
//var TestUrlRoot = "https://amarburg.github.io/go-lazyfs-testfiles/"
//var TestUrlRoot = "http://localhost:8080/files/"
//var TestUrl,_ = url.Parse( TestUrlRoot + TestMovPath )
var TestMovPath = "CamHD_Vent_Short.mov"


// For local testing
//var TestMovPath = lazyfs_testfiles.TestMovPath


var SparseHttpStoreRoot = "cache/httpsparse/"

func TestConvert( t *testing.T ) {

  srv := lazyfs_testfiles.HttpServer( 4567 )

  testUrl,err := url.Parse( srv.Url + TestMovPath )

  source,err := lazyfs.OpenHttpSource( *testUrl )
  //fmt.Println(source)
  //source,err := lazyfs.OpenLocalFileSource( "../go-lazyfs-testfiles/", TestMovPath )
  if err != nil {
    panic("Couldn't open HttpFSSource")
  }

  store,err := lazyfs.OpenSparseFileStore( source, SparseHttpStoreRoot )
  if store == nil {
    panic("Couldn't open SparesFileFSStore")
  }

  mov := LoadMovMetadata( store )

  // Try extracting a frame
  frame := 2
  img,_ := mov.ExtractFrame( frame )

  if err != nil { panic(fmt.Sprintf("Error decoding frame: %s", err.Error()))}

  img_filename := fmt.Sprintf("frame%06d.png", frame)
  img_file,err := os.Create(img_filename)
  if err != nil { panic(fmt.Sprintf("Error creating png %s: %s", img_filename, err.Error()))}

  err = png.Encode( img_file, img )
  if err != nil { panic(fmt.Sprintf("Error writing png %s: %s", img_filename, err.Error()))}


  srv.Stop()
}
