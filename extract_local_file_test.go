package lazyquicktime

import "testing"
import "fmt"

//import "io"
import "os"

import "github.com/amarburg/go-lazyfs"

import "image"
import "image/png"
import "github.com/amarburg/go-lazyfs-testfiles"
// import "github.com/amarburg/go-lazyfs-testfiles/http_server"


func doExtractFrame( t *testing.T, src lazyfs.FileSource ) (image.Image,error) {
	mov, _ := LoadMovMetadata(src)

	if( mov.NumFrames() != 31 ) {
		t.Errorf("Movie has incorrect number of frames (%d)", mov.NumFrames())
	}

	//fmt.Println("Movie has", mov.NumFrames(), "frames and is ", mov.Duration(), " seconds long")

	// Try extracting a frame
	frame := 2
	img, err := mov.ExtractFrame(frame)

	if err != nil {
		t.Errorf("Error decoding frame: %s", err.Error())
	}

	return img, err
}

func TestExtractFrameLocalFileSource(t *testing.T) {

	fileSource, err := lazyfs.OpenLocalFileSource(lazyfs_testfiles.RepoRoot(), lazyfs_testfiles.TestMovFile)

	if err != nil {
		panic("Couldn't open HttpFSSource")
	}

	_,err = doExtractFrame( t, fileSource )

	if err != nil {
		t.Errorf("Error extracting frame: %s", err.Error() )
	}

}

func TestSavePngFromLocalFileSource(t *testing.T) {

	// srv := lazyfs_testfiles_http_server.HttpServer()
	// defer srv.Stop()
	//
	// testUrl, err := url.Parse(srv.Url + lazyfs_testfiles.TestMovFile)
	//
	// source, err := lazyfs.OpenHttpSource(*testUrl)

	fileSource, err := lazyfs.OpenLocalFileSource(lazyfs_testfiles.RepoRoot(), lazyfs_testfiles.TestMovFile)

	if err != nil {
		t.Error("Couldn't open LocalFileSource")
	}

	img,_ := doExtractFrame( t, fileSource )

	img_filename := fmt.Sprintf("extracted_local.png")
	img_file, err := os.Create(img_filename)
	if err != nil {
		t.Errorf("Error creating png %s: %s", img_filename, err.Error())
	}

	err = png.Encode(img_file, img)
	if err != nil {
		t.Errorf("Error writing png %s: %s", img_filename, err.Error())
	}

}
