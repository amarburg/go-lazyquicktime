package lazyquicktime

import "testing"
import "fmt"


import "github.com/amarburg/go-lazyfs"
import "github.com/amarburg/go-lazyquicktime"


//import "net/url"
//var TestUrlRoot = "https://amarburg.github.io/go-lazyfs-testfiles/"
//var TestUrlRoot = "http://localhost:8080/files/"
//var TestUrl,_ = url.Parse( TestUrlRoot + TestMovPath )
//var TestMovPath = "CamHD_Vent_Short.mov"

// For local testing
import "github.com/amarburg/go-lazyfs-testfiles/http_server"
var TestMovPath = lazyfs_testfiles.TestMovPath

var SparseHttpStoreRoot = "cache/httpsparse/"

func BenchmarkExtractRepeatedFrameFromLocalSourceSparseStore( b *testing.B ) {

  //source,err := lazyfs.OpenHttpSource( *TestUrl )
  source,err := lazyfs.OpenLocalFileSource( "../go-lazyfs-testfiles/", TestMovPath )
  if err != nil { panic("Couldn't open HttpFSSource") }

  store,err := lazyfs.OpenSparseFileStore( source, SparseHttpStoreRoot )
  if store == nil {
    panic("Couldn't open SparesFileFSStore")
  }

  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    mov := lazyquicktime.LoadMovMetadata( store )

    // Try extracting a frame
    frame := 2
    mov.ExtractFrame( frame )

    if err != nil { panic(fmt.Sprintf("Error decoding frame: %s", err.Error()))}
  }


}
