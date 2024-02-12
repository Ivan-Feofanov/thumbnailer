package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/h2non/bimg"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err)
	}
	entries, err := os.ReadDir(cwd)
	if err != nil {
		log.Fatal().Err(err)
	}
	for _, file := range entries {
		buffer, err := bimg.Read(file.Name())
		if err != nil {
			log.Fatal().Err(err)
		}
		img := bimg.NewImage(buffer)
		rotated, err := img.AutoRotate()
		if err := bimg.Write(file.Name(), rotated); err != nil {
			log.Fatal().Err(err)
		}

		size, err := img.Size()
		if err != nil {
			log.Fatal().Err(err)
		}

		var div int
		if size.Width > 1000 || size.Height > 1000 {
			div = 4
		} else {
			div = 2
		}

		thumb, err := img.ResizeAndCrop(size.Width/div, size.Height/div)
		if err != nil {
			log.Fatal().Err(err)
		}

		if err = os.Mkdir("thumbs", 0755); err != nil {
			log.Fatal().Err(err)
		}
		if err := bimg.Write("thumbs/"+file.Name(), thumb); err != nil {
			log.Fatal().Err(err)
		}
	}
}
