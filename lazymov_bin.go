package main

import "fmt"
import "io"
import "os"

import "github.com/amarburg/go-lazyfs"
import "github.com/amarburg/go-quicktime"
import "github.com/amarburg/goav/avcodec"
import "github.com/amarburg/goav/avutil"
import "github.com/amarburg/goav/swscale"

import "image"
import "image/png"

import "encoding/binary"
import "bytes"

// //#cgo pkg-config: libavcodec
// //#include <libavcodec/avcodec.h>

import "C"
import "unsafe"


var TestUrlRoot = "http://localhost:8080/files/"
var TestMovPath = "CamHD_Vent_Short.mov"

var SparseHttpStoreRoot = "cache/httpsparse/"

func main() {

  source,err := lazyfs.OpenHttpFSSource(TestUrlRoot)
  if err != nil {
    panic("Couldn't open HttpFSSource")
  }

  store,err := lazyfs.OpenSparseFileFSStore( SparseHttpStoreRoot )
  if store == nil {
    panic("Couldn't open SparesFileFSStore")
  }

  source.SetBackingStore( store )

  file,err := source.Open( TestMovPath )
  if err != nil {
    panic("Couldn't open AlphabetPath")
  }

  // var offset int64 = 0
  // var indent = 0

  sz,_ := file.FileSize()
//  ParseAtom( file, offset, sz, indent )

  tree := quicktime.BuildTree( file, sz )

  quicktime.DumpTree( file, tree )

  moov := tree.FindAtom("moov")
  if moov == nil { panic("Can't find MOOV atom")}

  tracks := moov.FindAtoms("trak")
  if tracks == nil || len(tracks) == 0 { panic("Couldn't find any TRAKs in the MOOV")}
  //fmt.Println("Found",len(tracks),"TRAK atoms")

  var track *quicktime.Atom = nil
  for i,t := range tracks {
    mdia := t.FindAtom("mdia")
    if mdia == nil {
      fmt.Println("No mdia track",i)
      continue
    }

    minf := mdia.FindAtom("minf")
    if minf == nil {
      fmt.Println("No minf track",i)
      continue
    }

    if minf.FindAtom("vmhd") != nil {
      track = t
      break
    }
  }

  if track == nil { panic("Couldn't identify the Video track")}

  stbl_atom := track.FindAtom("mdia").FindAtom("minf").FindAtom("stbl")
  stbl,_ := quicktime.ParseSTBL( stbl_atom )

  //fmt.Println("Found track with video information")

  // Find movie length
  num_frames := stbl.NumFrames()

  fmt.Println("Movie has",num_frames,"frames")

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
  frame := 1
  LoadFrame( frame, stbl, file )
}


func LoadFrame( frame int, stbl quicktime.STBLAtom, file io.ReaderAt ) {

  frame_offset,frame_size,_ := stbl.SampleOffsetSize( frame )

  fmt.Printf("Extracting frame %d at offset %d size %d\n", frame, frame_offset, frame_size)

  buf := make([]byte, frame_size)
  n,_ := file.ReadAt( buf, frame_offset )

  if n != frame_size { panic(fmt.Sprintf("Tried to read %d bytes but got %d instead",frame_size,n))}

  avcodec.AvcodecRegisterAll()
  prores := avcodec.AvcodecFindDecoder( avcodec.CodecId(avcodec.AV_CODEC_ID_PRORES) )
  if prores == nil { panic("Couldn't find ProRes codec")}

  if prores.AvCodecIsDecoder() != 1 { panic("Isn't a decoder")}


  ctx := prores.AvcodecAllocContext3()
  if ctx == nil { panic("Couldn't allocate context") }

  res := ctx.AvcodecOpen2(prores,nil)
  if res < 0 { panic(fmt.Sprintf("Couldn't open context (%d)",res))}

  packet := avcodec.AvPacketAlloc()
  packet.AvInitPacket()
  //if packet == nil { panic("Couldn't allocate packet") }

  res = packet.AvPacketFromData( (*uint8)(unsafe.Pointer(&buf[0])), len(buf) )
  if res < 0 { panic(fmt.Sprintf("Couldn't set avpacket data (%d)",res))}

  // And a frame to receive the data
  videoFrame := avutil.AvFrameAlloc()
  if videoFrame == nil { panic("Couldn't allocate destination frame") }

  ctx.Width = 1920
  ctx.Height = 1080

  got_picture := 0
  res  = ctx.AvcodecDecodeVideo2( (*avcodec.Frame)(unsafe.Pointer(videoFrame)), &got_picture, packet )

fmt.Printf("Got picture: %d\n", got_picture)
fmt.Printf("%#v\n",*videoFrame)

  width,height := 1920,1080 //videoFrame.width, videoFrame.height

  if got_picture == 0 { panic(fmt.Sprintf("Didn't get a picture, err = %04x", -res)) }

  fmt.Printf("Image is %d x %d, format %d\n", width, height, int(ctx.Pix_fmt) )

  // Convert frame to RGB
  dest_fmt := int32(avcodec.AV_PIX_FMT_RGBA)
  flags := 0
  ctxtSws := swscale.SwsGetcontext(width, height, swscale.PixelFormat(ctx.Pix_fmt),
                                  width, height, swscale.PixelFormat(dest_fmt),
                                  flags, nil, nil, nil )
  if ctxtSws == nil { panic("Couldn't open swscale context")}

  videoFrameRgb := avutil.AvFrameAlloc()
  if videoFrameRgb == nil { panic("Couldn't allocate destination frame") }

  videoFrameRgb.Width = ctx.Width
  videoFrameRgb.Height = ctx.Height
  videoFrameRgb.Format = dest_fmt

  fmt.Println(videoFrameRgb )
  avutil.AvFrameGetBuffer( videoFrameRgb, 8)
  fmt.Println(videoFrameRgb )


  swscale.SwsScale( ctxtSws, videoFrame.Data,
                             videoFrame.Linesize,
                             0, height,
                             videoFrameRgb.Data,
                             videoFrameRgb.Linesize )

fmt.Printf("%#v\n",*videoFrameRgb)

  // Convert videoFrameRgb to Go image?
  img := image.NewRGBA( image.Rect(0,0,width,height) )

  img_filename := fmt.Sprintf("frame%06d.png", frame)
  img_file,err := os.Create(img_filename)
  if err != nil { panic(fmt.Sprintf("Error creating png %s: %s", img_filename, err.Error()))}

  rgb_data :=  C.GoBytes(unsafe.Pointer(videoFrameRgb.Data[0]), C.int(videoFrameRgb.Width * videoFrameRgb.Height * 4))
  //pixels := make([]byte, videoFrameRgb.Width * videoFrameRgb.Height * 4 )

  reader := bytes.NewReader( rgb_data )
  err = binary.Read( reader, binary.LittleEndian, &img.Pix)

  if err != nil { panic(fmt.Sprintf("error on binary read: %s", err.Error() ))}

  err = png.Encode( img_file, img )
  if err != nil { panic(fmt.Sprintf("Error writing png %s: %s", img_filename, err.Error()))}

}
