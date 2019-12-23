# m3d-dump

A program that reads through every `.M3D` 3D model file in Dark Omen's data and dumps out each of the model's textures and objects as JSON files.

## Installation

Use `go get` to download and install the program.

```shell
go get github.com/jonathaningram/dark-omen/cmd/m3d-dump
```

See `go help get` for more information.

## Usage

Pass the path to the Dark Omen CD data as well as the path to the output directory when running the program.

For Unix-based systems the command may look something like:

```shell
m3d-dump -dark-omen-path=/dark-omen-game-from-cd -output-path=/tmp/dark-omen-m3d-dump
```

For Windows the command may look something like:

```shell
m3d-dump.exe -dark-omen-path=D:\ -output-path=C:\tmp\dark-omen-m3d-dump
```

The output will look something like this (for a model file named `B1_01/BASE.M3D`):

```shell
$ ls -l /tmp/dark-omen-m3d-dump/DARKOMEN/DARKOMEN/GAMEDATA/1PBAT/B1_01/BASE.M3D/
object-0.json
object-1.json
object-2.json
object-3.json
texture-0.json
texture-1.json
texture-2.json
...
texture-34.json
texture-35.json
texture-36.json
```
