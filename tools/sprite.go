// +build ignore

package main

import (
	"image"
	"image/draw"
	"image/png"
	"os"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var frames = []string{
	"run1",
	"run2",
	"flap1",
	"flap2",
	"dead1",
	"dead2",
	"grass1",
	"grass2",
	"grass3",
	"grass4",
	"earth",
}

func main() {
	m := image.NewRGBA(image.Rect(0, 0, 128*len(frames), 128))
	for i, name := range frames {
		sm := readImage(name + ".png")
		draw.Draw(m, image.Rect(128*i, 0, 128*(i+1), 128), sm, image.ZP, draw.Over)
	}
	f, err := os.Create("sprite.png")
	check(err)
	check(png.Encode(f, m))
	check(f.Close())
}

func readImage(name string) image.Image {
	f, err := os.Open(name)
	check(err)
	defer f.Close()
	m, err := png.Decode(f)
	check(err)
	return m
}
