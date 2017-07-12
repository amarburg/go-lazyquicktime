package lazyquicktime

import "testing"

import "github.com/amarburg/go-lazyfs"

// Provides test files
import "github.com/amarburg/go-lazyfs-testfiles"

func extractFrameBenchmark(b *testing.B, source lazyfs.FileSource, f frameFunc) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mov, _ := LoadMovMetadata(source)

		// Try extracting a frame
		img, err := mov.ExtractFrame(f())

		if err != nil {
			b.Errorf("Error decoding frame: %s", err.Error())
		}

		if img.Bounds().Dx() != 1920 || img.Bounds().Dy() != 1080 {
			b.Errorf("Extracted frame wrong size (%d x %d)", img.Bounds().Dx(), img.Bounds().Dy())
		}
	}
}

func BenchmarkExtractFrameFromLocalSource(b *testing.B) {
	fileSource, err := lazyfs.OpenLocalFileSource(lazyfs_testfiles.RepoRoot(), lazyfs_testfiles.TestMovFile)
	if fileSource == nil || err != nil {
		panic("Couldn't open Test file")
	}

	extractFrameBenchmark(b, fileSource, func() int { return 2 })
}

func BenchmarkExtractFrameFromLocalSourceSparseStore(b *testing.B) {

	//source,err := lazyfs.OpenHttpSource( *TestUrl )
	fileSource, err := lazyfs.OpenLocalFileSource(lazyfs_testfiles.RepoRoot(), lazyfs_testfiles.TestMovFile)
	if err != nil {
		panic("Couldn't open Test file")
	}

	sparseStore, err := lazyfs.OpenSparseFileStore(fileSource, SparseStoreRoot)
	if sparseStore == nil || err != nil {
		panic("Couldn't open SparesFileFSStore")
	}

	extractFrameBenchmark(b, sparseStore, func() int { return 2 })
}
