# sad-dump

A program that reads through every `.SAD` stereo audio file in Dark Omen's data and dumps it out as a `.WAV` file.

## Installation

Use `go get` to download and install the program.

```shell
go get github.com/jonathaningram/dark-omen/cmd/sad-dump
```

See `go help get` for more information.

## Usage

Pass the path to the Dark Omen CD data as well as the path to the output directory when running the program.

For Unix-based systems the command may look something like:

```shell
sad-dump -dark-omen-path=/dark-omen-game-from-cd -output-path=/tmp/dark-omen-sad-dump
```

For Windows the command may look something like:

```shell
sad-dump.exe -dark-omen-path=D:\ -output-path=C:\tmp\dark-omen-sad-dump
```

The output will look something like this:

```shell
$ ls -l /tmp/dark-omen-sad-dump/DARKOMEN/DARKOMEN/SOUND/MUSIC/
1BOUN001.WAV
1BOUN002.WAV
1BOUN003.WAV
1BOUN004.WAV
...
12FOR002.WAV
12FOR003.WAV
12FOR004.WAV
SILENCE.WAV
```
