package main

import (
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
	// f, err := os.Open(filename)
	// if err != nil {
	// 	return errors.Wrap(err, "cannot open image file for embedding")
	// }
	// defer f.Close()
	// m, _, err := image.Decode(f)
	// if err != nil {
	// 	return errors.Wrap(err, "cannot decode image for embedding")
	// }

	// max is the resulting long side
	max := opts.width
	if opts.height > max {
		max = opts.height
	}
	max = int(float64(max) * opts.scale / 100)

	// // resize
	// filter := gift.New(gift.ResizeToFit(max, max, gift.LanczosResampling))
	// resized := image.NewRGBA(filter.Bounds(m.Bounds()))
	// filter.Draw(resized, m)
	w := int(float64(opts.width) * opts.scale / 100)
	h := int(float64(opts.height) * opts.scale / 100)

	resized := image.NewRGBA(image.Rect(0, 0, w, h))
	r := uint8(100 + rand.Intn(56))
	g := uint8(100 + rand.Intn(56))
	b := uint8(100 + rand.Intn(56))
	draw.Draw(resized, resized.Bounds(), image.NewUniform(color.RGBA{r, g, b, 255}), image.Point{}, draw.Src)

	// draw frame
	var (
		colorFrame = image.NewUniform(color.NRGBA{R: 242, G: 242, B: 242, A: 255})
		colorBlack = image.NewUniform(color.RGBA{0, 0, 0, 255})
	)
	border := int(opts.border / 100.0 * float64(max))
	p := image.Point{border, border}
	framed := image.NewNRGBA(resized.Bounds().Inset(-border).Add(p))
	draw.Draw(framed, framed.Bounds(), colorFrame, image.Point{}, draw.Src)
	draw.Draw(framed, resized.Bounds().Add(p), resized, image.Point{}, draw.Src)

	// draw shadow mask
	dropshadow := int(opts.dropshadow / 100.0 * float64(max))
	shadowShift := int(opts.dropshadow / 100.0 * float64(max) * 0.66)
	p = image.Point{dropshadow, shadowShift}
	shadow := image.NewRGBA(framed.Bounds().Inset(-dropshadow).Add(p))
	p = image.Point{shadowShift, shadowShift}
	draw.Draw(shadow, framed.Bounds().Add(p), colorBlack, image.Point{}, draw.Src)

	// blur shadow mask
	filter := gift.New(gift.GaussianBlur(float32(dropshadow) / 3))
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
