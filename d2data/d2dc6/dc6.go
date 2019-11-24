package d2dc6

import (
	"encoding/binary"
	"github.com/OpenDiablo2/D2Shared/d2data/d2datadict"
	"github.com/go-restruct/restruct"
	"github.com/hajimehoshi/ebiten"
	"log"
)

type DC6File struct {
	// Header
	Version            int32  `struct:"int32"`
	Flags              uint32 `struct:"uint32"`
	Encoding           uint32 `struct:"uint32"`
	Termination        []byte `struct:"[4]byte"`
	Directions         uint32 `struct:"uint32"`
	FramesPerDirection uint32 `struct:"uint32"`

	FramePointers []uint32    `struct:"[]uint32,size=Directions*FramesPerDirection"`
	Frames        []*DC6Frame `struct-size:"Directions*FramesPerDirection"`
	valid         bool
}

type DC6Frame struct {
	Flipped   uint32 `struct:"uint32"`
	Width     uint32 `struct:"uint32"`
	Height    uint32 `struct:"uint32"`
	OffsetX   int32  `struct:"int32"`
	OffsetY   int32  `struct:"int32"`
	Unknown   uint32 `struct:"uint32"`
	NextBlock uint32 `struct:"uint32"`
	Length    uint32 `struct:"uint32,sizeof=FrameData"`

	FrameData  []byte
	Terminator []byte `struct:"[3]byte"`

	imageData []int16
	image     *ebiten.Image
	palette   d2datadict.PaletteRec
	valid     bool
}

func (frame *DC6Frame) ImageData() []int16 {
	if frame.image == nil {
		frame.renderImage()
	}
	return frame.imageData
}

func (frame *DC6Frame) Image() *ebiten.Image {
	// Allows frames to be rendered lazily.
	if frame.image == nil {
		frame.renderImage()
	}
	return frame.image
}

func (frame *DC6Frame) renderImage() {
	frame.valid = true

	imageData := make([]int16, frame.Width*frame.Height)
	for fi := range imageData {
		imageData[fi] = -1
	}

	x := uint32(0)
	y := frame.Height - 1
	dataPointer := 0
	for {
		b := frame.FrameData[dataPointer]
		dataPointer++
		if b == 0x80 {
			if y == 0 {
				break
			}
			y--
			x = 0
		} else if (b & 0x80) > 0 {
			transparentPixels := b & 0x7F
			for ti := byte(0); ti < transparentPixels; ti++ {
				imageData[x+(y*frame.Width)+uint32(ti)] = -1
			}
			x += uint32(transparentPixels)
		} else {
			for bi := 0; bi < int(b); bi++ {
				imageData[x+(y*frame.Width)+uint32(bi)] = int16(frame.FrameData[dataPointer])
				dataPointer++
			}
			x += uint32(b)
		}
	}
	var img = make([]byte, int(frame.Width*frame.Height)*4)
	for ii := uint32(0); ii < frame.Width*frame.Height; ii++ {
		if imageData[ii] < 1 { // TODO: Is this == -1 or < 1?
			continue
		}
		img[ii*4] = frame.palette.Colors[imageData[ii]].R
		img[(ii*4)+1] = frame.palette.Colors[imageData[ii]].G
		img[(ii*4)+2] = frame.palette.Colors[imageData[ii]].B
		img[(ii*4)+3] = 0xFF
	}
	newImage, _ := ebiten.NewImage(int(frame.Width), int(frame.Height), ebiten.FilterNearest)
	err := newImage.ReplacePixels(img)
	if err != nil {
		log.Printf("failed to replace pixels: %v", err)
		frame.valid = false
	}

	frame.image = newImage
	frame.imageData = imageData
	// Probably don't need this data again
	frame.FrameData = nil
}

// LoadDC6 uses restruct to read the binary dc6 data into structs then parses image data from the frame data.
func LoadDC6(data []byte, palette d2datadict.PaletteRec) (DC6File, error) {
	result := DC6File{valid: true}

	restruct.EnableExprBeta()
	err := restruct.Unpack(data, binary.LittleEndian, &result)
	if err != nil {
		result.valid = false
		log.Printf("failed to read dc6: %v", err)
	}

	for _, frame := range result.Frames {
		frame.palette = palette
	}

	return result, err
}
