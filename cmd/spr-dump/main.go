package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jonathaningram/dark-omen/encoding/spr"
)

func main() {
	const (
		flagDarkOmenPath = "dark-omen-path"
		flagOutputPath   = "output-path"
	)

	var (
		darkOmenPath = flag.String(flagDarkOmenPath, "", "path to Dark Omen CD data")
		outputPath   = flag.String(flagOutputPath, "", "path to directory in which sprites will be dumped")
	)

	flag.Parse()

	if *darkOmenPath == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *outputPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	err := filepath.Walk(*darkOmenPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.ToUpper(path.Ext(info.Name())) != ".SPR" {
			return nil
		}

		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()

		relativePath := strings.TrimPrefix(p, *darkOmenPath)

		fmt.Printf("Decoding %s...", relativePath)

		d := spr.NewDecoder(f)

		sprite, err := d.Decode()
		if err != nil {
			fmt.Printf("failed\n")
			return fmt.Errorf("could not decode %s: %w", relativePath, err)
		}

		fmt.Printf("ok\n")

		dir := path.Join(*outputPath, relativePath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}

		fmt.Printf("Creating %d sprite(s) for %s...", len(sprite.Frames), relativePath)

		for i, f := range sprite.Frames {
			if f.Type == spr.FrameTypeEmpty {
				continue
			}

			file := path.Join(dir, fmt.Sprintf("%d.png", i))
			out, err := os.Create(file)
			if err != nil {
				fmt.Printf("failed\n")
				return err
			}
			defer out.Close()

			err = png.Encode(out, f.Image)
			if err != nil {
				fmt.Printf("failed\n")
				return fmt.Errorf("could not encode PNG file for frame %d: %w", i, err)
			}

			if err := out.Sync(); err != nil {
				return fmt.Errorf("could not sync PNG file for frame %d: %w", i, err)
			}
		}

		fmt.Printf("ok\n")

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
