package lazyquicktime

import "testing"

import "github.com/amarburg/go-lazyfs"

// For local testing
import "github.com/amarburg/go-lazyfs-testfiles"

var SparseStoreRoot = "cache/sparse/"

type frameFunc func() int

func loadMovMetadataTask( b *testing.B, source lazyfs.FileSource, f frameFunc ) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mov, err := LoadMovMetadata(source)

		if err != nil {
			b.Errorf("Error decoding frame: %s", err.Error())
		}

		// Assert the metadata is correct
		if( mov.NumFrames() != 31 ) {
			b.Errorf("Incorrect number of frames (%d)", mov.NumFrames())
		}
	}
}


func BenchmarkLoadMovMetadataFromLocalSource(b *testing.B) {

	fileSource, err := lazyfs.OpenLocalFileSource(lazyfs_testfiles.RepoRoot(), lazyfs_testfiles.TestMovFile)
	if fileSource == nil || err != nil {
		panic("Couldn't open Test file")
	}

	loadMovMetadataTask( b, fileSource, func() int { return 2 } )
}


func BenchmarkLoadMovMetadataFromLocalSourceSparseStore(b *testing.B) {

	//source,err := lazyfs.OpenHttpSource( *TestUrl )
	fileSource, err := lazyfs.OpenLocalFileSource(lazyfs_testfiles.RepoRoot(), lazyfs_testfiles.TestMovFile)
	if err != nil {
		panic("Couldn't open Test file")
	}

	sparseStore, err := lazyfs.OpenSparseFileStore(fileSource, SparseStoreRoot)
	if sparseStore == nil || err != nil {
		panic("Couldn't open SparesFileFSStore")
	}

	loadMovMetadataTask( b, sparseStore, func() int { return 2 } )
}
