package main

import (
	"flag"
	"fmt"
	"github.com/amarburg/go-lazyfs"
	"github.com/amarburg/go-lazyquicktime"
	"github.com/amarburg/go-quicktime"
	"image/png"
	"net/url"
	"os"
)

func main() {
	var frameNum, startAt, endAt int
	var srcURL, srcFile string
	var outputFile, statisticsFile string
	flag.IntVar(&frameNum, "frame", 0, "Frame number to extract")

	flag.IntVar(&startAt, "start", -1, "")
	flag.IntVar(&endAt, "end", -1, "")

	flag.StringVar(&srcURL, "url", "", "URL to query")
	flag.StringVar(&srcFile, "file", "", "URL to query")
	flag.StringVar(&outputFile, "output", "image_%09d.png", "Outputfile")
	flag.StringVar(&statisticsFile, "statistics", "", "Statistics file")

	flag.Parse()

	if len(srcURL) <= 0 && len(srcFile) <= 0 {
		panic("--url or --file must be specified")
	}

	var source lazyfs.FileSource
	var err error

	if len(srcURL) > 0 {
		testURL, err := url.Parse(srcURL)
		source, err = lazyfs.OpenHttpSource(*testURL)

		if err != nil {
			panic(fmt.Sprintf("Couldn't open HttpFSSource: %s", err.Error()))
		}
	} else if len(srcFile) > 0 {
		source, err = lazyfs.OpenLocalFileSource("./", srcFile)

		if err != nil {
			panic(fmt.Sprintf("Couldn't open LocalFileSource: %s", err.Error()))
		}
	}

	mov, _ := lazyquicktime.LoadMovMetadata(source)
	quicktime.DumpTree(mov.Tree)

	fmt.Println("Movie has", mov.NumFrames(), "frames and is ", mov.Duration(), " seconds long")

	if startAt < 0 && endAt < 0 {
		if frameNum > 0 {
			startAt = frameNum
			endAt = frameNum
		} else {
			panic("Neither -start and -end nor -frame were set")
		}
	}

	var statsFile *os.File
	if len(statisticsFile) > 0 {
		statsFile, err = os.Create(statisticsFile)
		if err != nil {
			panic(fmt.Sprintf("Unable to open stats file: %s", err.Error()))
		}
	}

	fmt.Printf("Processing from %d to %d\n", startAt, endAt)

	for i := startAt; i < endAt; i++ {

		frameOffset, frameSize, _ := mov.Stbl.SampleOffsetSize(i)
		chunk, chunkStart, remainder, err := mov.Stbl.Stsc.SampleChunk(i)

		if statsFile != nil {
			fmt.Fprintf(statsFile, "%d,%d,%d,%d,%d,%d,", i, frameOffset, frameSize, chunk, chunkStart, remainder)
		}

		// Try extracting a frame
		img, err := mov.ExtractFrame(uint64(i))

		if err != nil {
			fmt.Printf("Error decoding frame: %s\n", err.Error())
			fmt.Fprintf(statsFile, "N\n")
			continue
		}

		outFile := fmt.Sprintf(outputFile, i)

		imgFile, err := os.Create(outFile)
		if err != nil {
			fmt.Printf("Error creating png %s: %s\n", outFile, err.Error())
			fmt.Fprintf(statsFile, "N\n")
			continue
		}

		err = png.Encode(imgFile, img)
		if err != nil {
			fmt.Printf("Error writing png %s: %s\n", outFile, err.Error())
			fmt.Fprintf(statsFile, "N\n")
			continue
		}

		fmt.Fprintf(statsFile, "Y\n")

	}

}
