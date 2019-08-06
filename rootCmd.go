package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Options describe the conversion parameters
type Options struct {
	maxAngle    int
	quality     int
	output      string
	recursively bool
	number      int
	width       int
	height      int
	embedSize   int
	border      int
	dropshadow  int
}

var cmdRoot = &cobra.Command{
	Use:   "collage [flags] FILE/DIR [FILE/DIR] ...",
	Short: "Collage is a generator for a randomized photo stack collage",
	Long: `A generator for photo collages that appear do be dropped on a 
			stack. It takes a list of names of image files and/or directories 
			with image files which are then compositied into a collage image 
			and saved to disk.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		f := cmd.PersistentFlags()
		options := Options{}
		options.maxAngle, _ = f.GetInt("max-angle")
		options.quality, _ = f.GetInt("quality")
		options.output, _ = f.GetString("output")
		options.recursively, _ = f.GetBool("recursively")
		options.number, _ = f.GetInt("number")

		size, _ := f.GetString("size")
		if size == "" {
			return errors.New("size parameter may not be empty")
		}
		wh := strings.Split(size, "x")
		if len(wh) != 2 {
			return errors.New("invalid value for size, use e.g. 800x400")
		}
		w, err := strconv.Atoi(wh[0])
		if err != nil {
			return errors.Wrap(err, "invalid width value")
		}
		h, err := strconv.Atoi(wh[1])
		if err != nil {
			return errors.Wrap(err, "invalid height value")
		}
		options.width, options.height = w, h

		// max is the resulting long side
		max := options.width
		if options.height > max {
			max = options.height
		}
		embedSizePercent, _ := f.GetFloat64("embedsize")
		if embedSizePercent < 0 || embedSizePercent > 100 {
			return fmt.Errorf("invalid embedsize percentage value: %f", embedSizePercent)
		}
		options.embedSize = int(float64(max)*embedSizePercent/100.0 + 0.5)

		borderPercent, _ := f.GetFloat64("border")
		if borderPercent < 0 || borderPercent > 100 {
			return fmt.Errorf("invalid border percentage value: %f", borderPercent)
		}
		options.border = int(float64(max)*embedSizePercent*borderPercent/10000.0 + 0.5)

		dropPercent, _ := f.GetFloat64("dropshadow")
		if dropPercent < 0 || dropPercent > 100 {
			return fmt.Errorf("invalid dropshadow percentage value: %f", dropPercent)
		}
		options.dropshadow = int(float64(max)*embedSizePercent*dropPercent/10000.0 + 0.5)

		fmt.Printf("%+v", options)
		if len(args) == 0 {
			return fmt.Errorf("no FILE/DIR argument provided")
		}
		return run(options, args)
	},
}

func init() {
	f := cmdRoot.PersistentFlags()
	f.StringP("size", "s", "1920x1080", "target canvas size in WIDTHxHEIGHT")
	f.IntP("max-angle", "a", 60, "maximum rotation angle")
	f.IntP("quality", "q", 90, "JPEG quality parameter for resulting image")
	f.StringP("output", "o", "collage.jpg", "resulting image file name")
	f.BoolP("recursively", "r", false, "scan directories recursively")
	f.IntP("number", "n", 150, "maximum number of photos to use (0 means 'use all')")
	f.Float64P("embedsize", "e", 10.0, "size of embedded image in percent of target canvas size")
	f.Float64P("border", "b", 3.0, "size of border in percent of embedded image size")
	f.Float64P("dropshadow", "d", 3.0, "size of dropshadow in percent of embedded image size")
}

func run(opts Options, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no FILE/DIR argument provided")
	}
	all, err := findAllImages(args, opts.recursively)
	if err != nil {
		return errors.Wrap(err, "cannot identify conversion sources")
	}
	// shuffle the arguments
	for i := range all {
		r := rand.Intn(i + 1)
		all[r], all[i] = all[i], all[r]
	}
	if opts.number == 0 {
		opts.number = len(all)
	} else if opts.number > len(all) {
		opts.number = len(all) - 1
	}
	return convert(opts, all[:opts.number])
}

func convert(opts Options, args []string) error {
	targetImage := image.NewNRGBA(image.Rect(0, 0, opts.width, opts.height))
	for x := 0; x < opts.width; x++ {
		for y := 0; y < opts.height; y++ {
			targetImage.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	for i, a := range args {
		fmt.Printf("%d/%d: %s\n", i+1, opts.number, a)
		err := embedImage(targetImage, a, opts)
		if err != nil {
			return errors.Wrap(err, "cannot embed image")
		}
	}
	w, err := os.Create(opts.output)
	if err != nil {
		return errors.Wrap(err, "cannot create output file")
	}
	defer w.Close()
	return jpeg.Encode(w, targetImage, &jpeg.Options{Quality: opts.quality})
}

func findAllImages(args []string, recursively bool) ([]string, error) {
	var result []string
	for _, a := range args {
		fi, err := os.Lstat(a)
		if err != nil {
			return result, errors.Wrapf(err, "cannot get info for conversion source")
		}
		if !fi.IsDir() {
			result = append(result, a)
			continue
		}
		inDir, err := findImagesInDir(a, recursively)
		if err != nil {
			return result, errors.Wrapf(err, "cannot scan dir %s", a)
		}
		result = append(result, inDir...)
	}
	return result, nil
}

func findImagesInDir(dir string, recursively bool) ([]string, error) {
	var result []string
	err := filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() && p != dir && !recursively {
			return filepath.SkipDir
		}
		switch strings.ToLower(filepath.Ext(p)) {
		default:
		case ".png", ".jpeg", ".jpg":
			result = append(result, p)
		}
		return nil
	})
	return result, err
}
