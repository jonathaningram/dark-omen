package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jonathaningram/dark-omen/encoding/sad"
)

func writeWAV(s *sad.Stream, dir, fileName string) error {
	file := path.Join(dir, fmt.Sprintf("%s.WAV", fileName))
	fmt.Printf("Creating WAV %s...", file)
	out, err := os.Create(file)
	if err != nil {
		fmt.Printf("failed\n")
		return err
	}
	defer out.Close()

	if err := s.EncodeToWAV(out); err != nil {
		fmt.Printf("failed\n")
		return fmt.Errorf("could not convert to WAV: %w", err)
	}

	if err := out.Sync(); err != nil {
		fmt.Printf("failed\n")
		return err
	}
	fmt.Printf("ok\n")
	return nil
}

func main() {
	const (
		flagDarkOmenPath = "dark-omen-path"
		flagOutputPath   = "output-path"
	)

	var (
		darkOmenPath = flag.String(flagDarkOmenPath, "", "path to Dark Omen CD data")
		outputPath   = flag.String(flagOutputPath, "", "path to directory in which WAV files will be dumped")
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
		if strings.ToUpper(path.Ext(info.Name())) != ".SAD" {
			return nil
		}

		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()

		relativePath := strings.TrimPrefix(p, *darkOmenPath)

		fmt.Printf("Decoding %s...", relativePath)

		stream, err := sad.NewDecoder(f).Decode()
		if err != nil {
			fmt.Printf("failed\n")
			return err
		}

		fmt.Printf("ok\n")

		dir := path.Join(*outputPath, path.Dir(relativePath))
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}

		if err := writeWAV(stream, dir, path.Base(strings.TrimSuffix(relativePath, path.Ext(relativePath)))); err != nil {
			return fmt.Errorf("could not write WAV: %w", err)
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
