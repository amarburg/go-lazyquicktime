package lazyquicktime

import "testing"
import "fmt"


import "github.com/amarburg/go-lazyfs"

// For local testing
import "github.com/amarburg/go-lazyfs-testfiles"

var SparseStoreRoot = "cache/sparse/"

func BenchmarkExtractRepeatedFrameFromLocalSourceSparseStore( b *testing.B ) {

  //source,err := lazyfs.OpenHttpSource( *TestUrl )
  source,err := lazyfs.OpenLocalFileSource( lazyfs_testfiles.RepoRoot(), lazyfs_testfiles.TestMovFile )
  if err != nil { panic("Couldn't open Test file") }

  store,err := lazyfs.OpenSparseFileStore( source, SparseStoreRoot )
  if store == nil {
    panic("Couldn't open SparesFileFSStore")
  }

  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    mov,_ := LoadMovMetadata( store )

    // Try extracting a frame
    frame := 2
    mov.ExtractFrame( frame )

    if err != nil { panic(fmt.Sprintf("Error decoding frame: %s", err.Error()))}
  }


}
