package lazyquicktime

import "testing"
import "fmt"
import "net/url"
//import "io"
import "os"

import "github.com/amarburg/go-lazyfs"

import "image/png"
import "github.com/amarburg/go-lazyfs-testfiles"
import "github.com/amarburg/go-lazyfs-testfiles/http_server"


func BenchmarkConvert( b *testing.B ) {

  srv := lazyfs_testfiles_http_server.HttpServer( )
  defer srv.Stop()

  testUrl,err := url.Parse( srv.Url + lazyfs_testfiles.TestMovFile )
  source,err := lazyfs.OpenHttpSource( *testUrl )
  if err != nil {
    panic("Couldn't open HttpFSSource")
  }

  store,err := lazyfs.OpenSparseFileStore( source, SparseHttpStoreRoot )
  if store == nil {
    panic("Couldn't open SparesFileFSStore")
  }

  mov := LoadMovMetadata( store )

  b.ResetTimer()
  for i := 0; i < b.N; i++ {

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
b.StopTimer()

}
