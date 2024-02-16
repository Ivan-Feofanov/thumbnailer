package main_test

import (
	"bytes"
	"image"
	"image/jpeg"
	"os"
	"testing"

	main "github.com/Ivan-Feofanov/thumbnailer"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/h2non/bimg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsImage(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{filename: "image.jpg", expected: true},
		{filename: "image.jpeg", expected: true},
		{filename: "image.png", expected: true},
		{filename: "image.gif", expected: true},
		{filename: "image.webp", expected: true},
		{filename: "document.pdf", expected: false},
		{filename: "video.mp4", expected: false},
		{filename: "text.txt", expected: false},
	}

	for _, test := range tests {
		actual := main.IsImage(test.filename)
		if actual != test.expected {
			t.Errorf("IsImage(%s) = %v, expected %v", test.filename, actual, test.expected)
		}
	}
}

type DirEntry struct {
	name     string
	isDir    bool
	thisType os.FileMode
	info     os.FileInfo
}

func (d *DirEntry) IsDir() bool {
	return d.isDir
}
func (d *DirEntry) Name() string {
	return d.name
}
func (d *DirEntry) Type() os.FileMode {
	return d.thisType
}
func (d *DirEntry) Info() (os.FileInfo, error) {
	return d.info, nil
}

func TestFilterImages(t *testing.T) {
	type args struct {
		assets []os.DirEntry
	}
	tests := []struct {
		name string
		args args
		want []os.DirEntry
	}{
		{
			name: "no images",
			args: args{
				assets: []os.DirEntry{
					&DirEntry{name: "document.pdf"},
					&DirEntry{name: "video.mp4"},
					&DirEntry{name: "text.txt"},
				},
			},
			want: []os.DirEntry(nil),
		},
		{
			name: "some images",
			args: args{
				assets: []os.DirEntry{
					&DirEntry{name: "image.jpg"},
					&DirEntry{name: "image.jpeg"},
					&DirEntry{name: "image.png"},
					&DirEntry{name: "image.gif"},
					&DirEntry{name: "image.webp"},
					&DirEntry{name: "document.pdf"},
					&DirEntry{name: "video.mp4"},
					&DirEntry{name: "text.txt"},
				},
			},
			want: []os.DirEntry{
				&DirEntry{name: "image.jpg"},
				&DirEntry{name: "image.jpeg"},
				&DirEntry{name: "image.png"},
				&DirEntry{name: "image.gif"},
				&DirEntry{name: "image.webp"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := main.FilterImages(tt.args.assets)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCreateThumbnail(t *testing.T) {
	type args struct {
		img *image.RGBA
	}

	smallImg := gofakeit.Image(1000, 1000)
	smallImageBuf := new(bytes.Buffer)
	require.NoError(t, jpeg.Encode(smallImageBuf, smallImg, nil))

	bigImageBuf := new(bytes.Buffer)
	bigImg := gofakeit.Image(2000, 2000)
	require.NoError(t, jpeg.Encode(bigImageBuf, bigImg, nil))

	tests := []struct {
		name    string
		args    args
		want    bimg.ImageSize
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "no error, small image",
			args: args{
				img: gofakeit.Image(1000, 1000),
			},
			want:    bimg.ImageSize{Width: 500, Height: 500},
			wantErr: assert.NoError,
		},
		{
			name: "no error, big image",
			args: args{
				img: gofakeit.Image(2000, 2000),
			},
			want:    bimg.ImageSize{Width: 500, Height: 500},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			require.NoError(t, jpeg.Encode(buf, tt.args.img, nil))
		
			got, err := main.CreateThumbnail(bimg.NewImage(buf.Bytes()))
	
			if tt.wantErr(t, err) {
				size, err := bimg.NewImage(got).Size()
				require.NoError(t, err)
				assert.Equal(t, tt.want, size)
			}
		})
	}
}
