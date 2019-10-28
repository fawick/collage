# Collage 

A generator for photo collages that appear do be dropped on a 
stack. It takes a list of names of image files and/or directories
with image files which are then compositied into a collage image
and saved to disk.

![collage.jpg](collage.jpg)


## Installation

Clone the repository and then run

```
$ GO111MODULES=1 go build
```

## Usage

```
collage [flags] FILE/DIR [FILE/DIR] ...

Flags:
  -b, --border float       size of border in percent of embedded image size (default 3)
  -d, --dropshadow float   size of dropshadow in percent of embedded image size (default 3)
  -e, --embedsize float    size of embedded image in percent of target canvas size (default 10)
  -h, --help               help for collage
  -a, --max-angle int      maximum rotation angle (default 60)
  -n, --number int         maximum number of photos to use (0 means 'use all') (default 150)
  -o, --output string      resulting image file name (default "collage.jpg")
  -q, --quality int        JPEG quality parameter for resulting image (default 90)
  -r, --recursively        scan directories recursively
  -s, --size string        target canvas size in WIDTHxHEIGHT (default "1920x1080")
```
