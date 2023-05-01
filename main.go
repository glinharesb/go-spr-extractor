package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"

	buffer_reader "github.com/glinharesb/go-buffer-reader"

	"github.com/cheggaaa/pb"
)

const (
	SPR_FILE = "Tibia.spr"
	OUT_DIR  = "./output"

	BASE_COLOR_R = 255
	BASE_COLOR_G = 0
	BASE_COLOR_B = 255
	BASE_COLOR_A = 255
	BASE_IMAGE_W = 32
	BASE_IMAGE_H = 32
)

func exportImage(obj interface{}) {
	data := obj.(map[string]interface{})
	filename := data["filename"].(string)
	img := data["img"].(*image.NRGBA)

	file, err := os.Create(filepath.Join(OUT_DIR, filename+".png"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func processSprite(spriteID int, reader *buffer_reader.BufferReader) {
	filename := strconv.Itoa(spriteID)
	img := image.NewNRGBA(image.Rect(0, 0, BASE_IMAGE_W, BASE_IMAGE_H))

	formula := 6 + (spriteID-1)*4
	reader.Seek(formula)

	address := reader.NextUInt32LE()
	if address == 0 {
		return
	}
	reader.Seek(int(address))
	reader.Move(3)

	offset := reader.Tell() + int(reader.NextUInt16LE())
	currentPixel := 0
	size := 32

	for reader.Tell() < offset {
		transparentPixels := reader.NextUInt16LE()
		coloredPixels := reader.NextUInt16LE()
		currentPixel += int(transparentPixels)

		for i := 0; i < int(coloredPixels); i++ {
			x := int(currentPixel % size)
			y := int(currentPixel / size)

			color := color.NRGBA{
				R: reader.NextUInt8(),
				G: reader.NextUInt8(),
				B: reader.NextUInt8(),
				A: 255,
			}

			img.Set(x, y, color)

			currentPixel++
		}
	}

	baseColor := color.NRGBA{
		R: BASE_COLOR_R,
		G: BASE_COLOR_G,
		B: BASE_COLOR_B,
		A: BASE_COLOR_A,
	}

	for x := 0; x < BASE_IMAGE_W; x++ {
		for y := 0; y < BASE_IMAGE_H; y++ {
			if img.At(x, y).(color.NRGBA).A == 0 {
				img.Set(x, y, baseColor)
			}
		}
	}

	exportImage(map[string]interface{}{"filename": filename, "img": img})
}

func processSprFile(buffer []byte) {
	reader := buffer_reader.NewBufferReader(buffer)

	reader.NextUInt32LE()
	size := int(reader.NextUInt16LE())

	bar := pb.New(size - 1).Prefix("Processing sprites ")
	bar.Start()

	for spriteID := 1; spriteID < size; spriteID++ {
		processSprite(spriteID, reader)
		bar.Set(spriteID)
	}

	bar.Finish()

	log.Printf("%d images exported successfully\n", size-1)
}

func createOutputDirIfNotExist() error {
	if _, err := os.Stat(OUT_DIR); os.IsNotExist(err) {
		err := os.Mkdir(OUT_DIR, 0755)
		if err != nil {
			return fmt.Errorf("error creating output directory: %s", err.Error())
		}
	}

	return nil
}

func main() {
	var err error

	err = createOutputDirIfNotExist()
	if err != nil {
		panic(err)
	}

	buffer, err := os.ReadFile(SPR_FILE)
	if err != nil {
		panic(err)
	}

	processSprFile(buffer)
}
