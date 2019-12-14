# spr-dump

A program that reads through every `.SPR` sprite file in Dark Omen's data and dumps out each of the sprite's frames as PNG images.

## Installation

Use `go get` to download and install the program.

```shell
go get github.com/jonathaningram/dark-omen/cmd/spr-dump
```

See `go help get` for more information.

## Usage

Pass the path to the Dark Omen CD data as well as the path to the output directory when running the program.

For Unix-based systems the command may look something like:

```shell
spr-dump -dark-omen-path=/dark-omen-game-from-cd -output-path=/tmp/dark-omen-spr-dump
```

For Windows the command may look something like:

```shell
spr-dump.exe -dark-omen-path=D:\ -output-path=C:\tmp\dark-omen-spr-dump
```

The output will look something like this (for a sprite file named `BERNHD.SPR`):

```shell
$ ls -l /tmp/dark-omen-spr-dump/DARKOMEN/DARKOMEN/GRAPHICS/SPRITES/BERNHD.SPR/
0.png
1.png
2.png
3.png
...
100.png
101.png
102.png
103.png
```
