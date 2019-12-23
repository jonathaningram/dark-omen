package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jonathaningram/dark-omen/encoding/m3d"
)

func writeObject(index int, o *m3d.Object, dir string) error {
	file := path.Join(dir, fmt.Sprintf("object-%d.json", index))
	out, err := os.Create(file)
	if err != nil {
		return err
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	enc.SetIndent("", "\t")
	err = enc.Encode(o)
	if err != nil {
		return fmt.Errorf("could not encode JSON file: %w", err)
	}

	return out.Sync()
}

func writeObjects(model *m3d.Model, relativePath, dir string) error {
	fmt.Printf("Creating %d object(s) for %s...", len(model.Objects), relativePath)

	for i, o := range model.Objects {
		if err := writeObject(i, o, dir); err != nil {
			fmt.Printf("\n...failed\n")
			return fmt.Errorf("could not write object %d: %w", i, err)
		}
		fmt.Printf("\n- %s", o.Name)
	}

	fmt.Printf("\n...ok\n")

	return nil
}

func writeTexture(index int, t *m3d.Texture, dir string) error {
	file := path.Join(dir, fmt.Sprintf("texture-%d.json", index))
	out, err := os.Create(file)
	if err != nil {
		return err
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	enc.SetIndent("", "\t")
	err = enc.Encode(t)
	if err != nil {
		return fmt.Errorf("could not encode JSON file: %w", err)
	}

	return out.Sync()
}

func writeTextures(model *m3d.Model, relativePath, dir string) error {
	fmt.Printf("Creating %d texture(s) for %s...", len(model.Textures), relativePath)

	for i, t := range model.Textures {
		if err := writeTexture(i, t, dir); err != nil {
			fmt.Printf("\n...failed\n")
			return fmt.Errorf("could not write texture %d: %w", i, err)
		}
		fmt.Printf("\n- %s", t.FileName)
	}

	fmt.Printf("\n...ok\n")

	return nil
}

func main() {
	const (
		flagDarkOmenPath = "dark-omen-path"
		flagOutputPath   = "output-path"
	)

	var (
		darkOmenPath = flag.String(flagDarkOmenPath, "", "path to Dark Omen CD data")
		outputPath   = flag.String(flagOutputPath, "", "path to directory in which models will be dumped")
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
		if strings.ToUpper(path.Ext(info.Name())) != ".M3D" {
			return nil
		}

		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()

		relativePath := strings.TrimPrefix(p, *darkOmenPath)

		fmt.Printf("Decoding %s...", relativePath)

		d := m3d.NewDecoder(f)

		model, err := d.Decode()
		if err != nil {
			fmt.Printf("failed\n")
			return fmt.Errorf("could not decode %s: %w", relativePath, err)
		}

		fmt.Printf("ok\n")

		dir := path.Join(*outputPath, relativePath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}

		if err := writeTextures(model, relativePath, dir); err != nil {
			return fmt.Errorf("could not write textures: %w", err)
		}

		if err := writeObjects(model, relativePath, dir); err != nil {
			return fmt.Errorf("could not write objects: %w", err)
		}

		return nil
	})
	check(err)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
