# mad-dump

A program that reads through every `.MAD` mono audio file in Dark Omen's data and dumps it out as a `.WAV` file.

## Installation

Use `go get` to download and install the program.

```shell
go get github.com/jonathaningram/dark-omen/cmd/mad-dump
```

See `go help get` for more information.

## Usage

Pass the path to the Dark Omen CD data as well as the path to the output directory when running the program.

For Unix-based systems the command may look something like:

```shell
mad-dump -dark-omen-path=/dark-omen-game-from-cd -output-path=/tmp/dark-omen-mad-dump
```

For Windows the command may look something like:

```shell
mad-dump.exe -dark-omen-path=D:\ -output-path=C:\tmp\dark-omen-mad-dump
```

The output will look something like this:

```shell
$ ls -l /tmp/dark-omen-mad-dump/DARKOMEN/DARKOMEN/SOUND/SP_ENG/
A_AARGH1.WAV
A_AARGH2.WAV
...
A_ENGAGE.WAV
...
WH058.WAV
WH059.WAV
```
