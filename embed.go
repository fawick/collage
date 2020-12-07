package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math/rand"
	"os"

	"github.com/disintegration/gift"
)

type rectMask struct {
	m image.Image
}

func embedImage(target draw.Image, filename string, opts Options) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open image file for embedding: %w", err)
	}
	defer f.Close()
	m, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("cannot decode image for embedding: %w", err)
	}

	// resize
	filter := gift.New(gift.ResizeToFit(opts.embedSize, opts.embedSize, gift.LanczosResampling))
	resized := image.NewRGBA(filter.Bounds(m.Bounds()))
	filter.Draw(resized, m)

	// draw frame
	var (
		colorFrame = image.NewUniform(color.NRGBA{R: 242, G: 242, B: 242, A: 255})
		colorBlack = image.NewUniform(color.RGBA{0, 0, 0, 64})
	)
	p := image.Point{opts.border, opts.border}
	framed := image.NewNRGBA(resized.Bounds().Inset(-opts.border).Add(p))
	draw.Draw(framed, framed.Bounds(), colorFrame, image.Point{}, draw.Src)
	draw.Draw(framed, resized.Bounds().Add(p), resized, image.Point{}, draw.Src)

	// draw shadow mask
	p = image.Point{opts.dropshadow, opts.dropshadow}
	shadow := image.NewRGBA(framed.Bounds().Inset(-opts.dropshadow).Add(p))
	draw.Draw(shadow, shadow.Bounds().Inset(opts.dropshadow/2), colorBlack, image.Point{}, draw.Src)

	// blur shadow mask
	filter = gift.New(gift.GaussianBlur(float32(opts.dropshadow) / 3.0))
	dst := image.NewRGBA(filter.Bounds(shadow.Bounds()))
	filter.Draw(dst, shadow)
	draw.Draw(dst, framed.Bounds().Add(p), framed, image.Point{}, draw.Src)

	// rotate by random angle
	a := rand.NormFloat64() * float64(opts.maxAngle) / 3
	filter = gift.New(gift.Rotate(float32(a), color.Transparent, gift.LinearInterpolation))
	rotated := image.NewRGBA(filter.Bounds(dst.Bounds()))
	filter.Draw(rotated, dst)

	//embed in image
	p = image.Point{
		X: rand.Intn(target.Bounds().Dx() - rotated.Bounds().Dx()),
		Y: rand.Intn(target.Bounds().Dy() - rotated.Bounds().Dy()),
	}
	draw.Draw(target, rotated.Bounds().Add(p), rotated, image.Point{}, draw.Over)

	return nil
}

func dummyWriteImage(filename string, m image.Image) {
	w, err := os.Create(filename)
	if err != nil {
		log.Println(err)
	}
	defer w.Close()
	if err = png.Encode(w, m); err != nil {
		log.Println(err)
	}
}
