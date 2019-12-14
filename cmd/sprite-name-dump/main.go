package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/jonathaningram/dark-omen/engrel"
)

func main() {
	const flagDarkOmenPath = "dark-omen-path"

	darkOmenPath := flag.String(flagDarkOmenPath, "", "path to Dark Omen CD data")

	flag.Parse()

	if *darkOmenPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.Open(path.Join(*darkOmenPath, "DARKOMEN", "DARKOMEN", "PRG_ENG", "ENGREL.EXE"))
	check(err)
	defer f.Close()

	names, err := engrel.ReadSpriteNames(f)
	check(err)

	for _, s := range names {
		fmt.Println(s)
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
