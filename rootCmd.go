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
	size        string
	output      string
	recursively bool
	number      int
	border      int
	width       int
	height      int
	dropshadow  int
}

var options Options

var cmdRoot = &cobra.Command{
	Use:   "collage [flags] FILE/DIR [FILE/DIR] ...",
	Short: "Collage is a generator for a randomized photo stack collage",
	Long: `A generator for photo collages that appear do be dropped on a 
			stack. It takes a list of names of image files and/or directories 
			with image files which are then compositied into a collage image 
			and saved to disk.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if options.size == "" {
			return errors.New("size parameter may not be empty")
		}
		wh := strings.Split(options.size, "x")
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
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no FILE/DIR argument provided")
		}
		return run(options, args)
	},
}

func init() {
	f := cmdRoot.PersistentFlags()
	f.StringVarP(&options.size, "size", "s", "1920x1080", "target canvas size in WIDTHxHEIGHT")
	f.IntVarP(&options.maxAngle, "max-angle", "a", 60, "maximum rotation angle")
	f.IntVarP(&options.quality, "quality", "q", 90, "JPEG quality parameter for resulting image")
	f.IntVarP(&options.number, "number", "n", 150, "maximum number of photos to use (0 means 'use all')")
	f.IntVarP(&options.border, "border", "b", 25, "border width in pixel")
	f.IntVarP(&options.dropshadow, "dropshadow", "d", 25, "drop shadow width in pixel")
	f.StringVarP(&options.output, "output", "o", "collage.jpg", "resulting image file name")
	f.BoolVarP(&options.recursively, "recursively", "r", false, "scan directories recursively")
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
	for i := range args {
		r := rand.Intn(i + 1)
		args[r], args[i] = args[i], args[r]
	}
	if opts.number == 0 {
		opts.number = len(all)
	} else if opts.number > len(all) {
		opts.number = len(all) - 1
	}
	return convert(opts, args[:opts.number])
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
