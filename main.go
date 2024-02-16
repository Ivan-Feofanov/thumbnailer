package main

import (
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/h2non/bimg"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
)

func IsImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	}
	return false
}

func FilterImages(assets []fs.DirEntry) []os.DirEntry {
	var images []os.DirEntry
	for _, asset := range assets {
		if asset.IsDir() || !IsImage(asset.Name()) {
			continue
		}
		images = append(images, asset)
	}
	return images
}

func CreateThumbnail(img *bimg.Image) ([]byte, error) {
	size, err := img.Size()
	if err != nil {
		return nil, err
	}

	var div int
	if size.Width > 1000 || size.Height > 1000 {
		div = 4
	} else {
		div = 2
	}

	thumb, err := img.ResizeAndCrop(size.Width/div, size.Height/div)
	if err != nil {
		return nil, err
	}
	return thumb, nil
}

func ScaleDown(img *bimg.Image, n int) ([]byte, error) {
	size, err := img.Size()
	if err != nil {
		return nil, err
	}

	scaled, err := img.Resize(size.Width/n, size.Height/n)
	if err != nil {
		return nil, err
	}
	return scaled, nil
}

func ConvertAndWrite(img *bimg.Image, filename string) error {
	imgBytes, err := img.Convert(bimg.WEBP)
	if err != nil {
		return err
	}
	if err := bimg.Write(filename+".webp", imgBytes); err != nil {
		return err
	}
	return nil
}

func main() {
	skipThumb := flag.Bool("no-thumbs", false, "do not generate thumbnails")
	scaleDownRate := flag.Int("scale-down", 1, "scale down original images")
	enumerate := flag.Bool("enumerate", false, "enumerate all images")
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err)
	}
	entries, err := os.ReadDir(cwd)
	if err != nil {
		log.Fatal().Err(err)
	}
	images := FilterImages(entries)

	bar := progressbar.Default(int64(len(images)))
	for idx, file := range images {
		buffer, err := bimg.Read(file.Name())
		if err != nil {
			log.Fatal().Err(err)
		}

		fileName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		if *enumerate {
			fileName = strconv.Itoa(idx)
		}

		img := bimg.NewImage(buffer)
		rotated, err := img.AutoRotate()
		if err != nil {
			log.Fatal().Err(err)
		}
		img = bimg.NewImage(rotated)
		imgSize, err := img.Size()
		if err != nil {
			log.Fatal().Err(err)
		}
		log.Info().Str("file", file.Name()).Int("width", imgSize.Width).Int("height", imgSize.Height).Msg("processing")
		if *scaleDownRate > 1 && (imgSize.Width > 1280 || imgSize.Height > 1280) {
			scaled, err := ScaleDown(img, *scaleDownRate)
			if err != nil {
				log.Fatal().Err(err)
			}
			img = bimg.NewImage(scaled)
		}

		if err := ConvertAndWrite(img, fileName); err != nil {
			log.Fatal().Err(err)
		}

		if !*skipThumb {
			thumb, err := CreateThumbnail(img)
			if err != nil {
				log.Fatal().Err(err)
			}
			if err = os.Mkdir("thumbs", 0755); err != nil {
				log.Fatal().Err(err)
			}
			if err := ConvertAndWrite(bimg.NewImage(thumb), "thumbs/"+fileName); err != nil {
				log.Fatal().Err(err)
			}
		}

		if err := bar.Add(1); err != nil {
			log.Fatal().Err(err)
		}
	}
}
