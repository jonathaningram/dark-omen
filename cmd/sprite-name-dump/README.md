# sprite-name-dump

A program that reads the Dark Omen sprite names from the `PRG_ENG/ENGREL.EXE` file and dumps them to stdout.

## Installation

Use `go get` to download and install the program.

```shell
go get github.com/jonathaningram/dark-omen/cmd/sprite-name-dump
```

See `go help get` for more information.

## Usage

Pass the path to the Dark Omen CD data when running the program.

For Unix-based systems the command may look something like:

```shell
sprite-name-dump -dark-omen-path=/dark-omen-game-from-cd
```

For Windows the command may look something like:

```shell
sprite-name-dump.exe -dark-omen-path=D:\
```

The output will look something like:

```shell
VOIDTYPE
BtlSprit
flags
missiles
mi
SPL_ITEM
SPL_BRI
SPL_BRI
SPL_BRI
SPL_DARK
XST_ZNewMisc5
XST_ZNewMisc6
GRAILK
KREALM
DWARF
BERNHD
...
DBGYPARC
BT_ZZNewUndeadBan4
BT_ZZNewUndeadBan5
BT_ZZNewUndeadBan6
```
