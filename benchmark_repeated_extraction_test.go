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

var SparseHTTPStoreRoot = "test_cache/httpsparse/"

func BenchmarkConvert(b *testing.B) {

	srv := lazyfs_testfiles_http_server.HttpServer()
	defer srv.Stop()

	testURL, err := url.Parse(srv.Url + lazyfs_testfiles.TestMovFile)
	source, err := lazyfs.OpenHttpSource(*testURL)
	if err != nil {
		panic("Couldn't open HttpFSSource")
	}

	store, err := lazyfs.OpenSparseFileStore(source, SparseHTTPStoreRoot)
	if store == nil {
		panic("Couldn't open SparesFileFSStore")
	}

	mov, _ := LoadMovMetadata(store)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		// Try extracting a frame
		frame := uint64(2)
		img, _ := mov.ExtractFrame(frame)

		if err != nil {
			panic(fmt.Sprintf("Error decoding frame: %s", err.Error()))
		}

		imgFilename := fmt.Sprintf("frame%06d.png", frame)
		imgFile, err := os.Create(imgFilename)
		if err != nil {
			panic(fmt.Sprintf("Error creating png %s: %s", imgFilename, err.Error()))
		}

		err = png.Encode(imgFile, img)
		if err != nil {
			panic(fmt.Sprintf("Error writing png %s: %s", imgFilename, err.Error()))
		}
	}
	b.StopTimer()

}
