# Warhammer: Dark Omen Go library

[![Go reference](https://pkg.go.dev/badge/github.com/jonathaningram/dark-omen.svg)](https://pkg.go.dev/github.com/jonathaningram/dark-omen)
[![Build status](https://github.com/jonathaningram/dark-omen/workflows/Go/badge.svg?branch=main)](https://github.com/jonathaningram/dark-omen/actions)
[![Report card](https://goreportcard.com/badge/github.com/jonathaningram/dark-omen)](https://goreportcard.com/report/github.com/jonathaningram/dark-omen)
[![Sourcegraph](https://sourcegraph.com/github.com/jonathaningram/dark-omen/-/badge.svg)](https://sourcegraph.com/github.com/jonathaningram/dark-omen)

This library is for developers interested in building tools for working with Dark Omen's assets.

This library does not ship with any Dark Omen assets. You must have a legally purchased copy of Dark Omen in order to get the benefits of this library.

**Note:** This library is neither developed by nor endorsed by Electronic Arts Inc.

## Table of contents

- [Installation](#installation)
- [Game file support](#game-file-support)
- [Tests](#tests)

## Installation

Use `go get` to retrieve the library and add it to your project's Go module dependencies.

```shell
go get github.com/jonathaningram/dark-omen
```

## Game file support

The following game files are supported by the library.

| Kind                                 | File extension | Read | Write | Known issues?                                   |
| ------------------------------------ | -------------- | ---- | ----- | ----------------------------------------------- |
| [Army and saved games](encoding/arm) | .ARM           | ✅   | ❌    | ⚠️ Yes, experimental and incomplete             |
| [Dot](encoding/dot)                  | .DOT           | ✅   | ❌    | ✅ None                                         |
| [Font](encoding/fnt)                 | .FNT           | ✅   | ❌    | ⚠️ Yes, height/line-height possibly not correct |
| [3D model](encoding/m3d)             | .M3D           | ✅   | ❌    | ✅ None                                         |
| [Mono audio](encoding/mad)           | .MAD           | ✅   | ✅    | ✅ None                                         |
| [Project](encoding/prj)              | .PRJ           | ✅   | ❌    | ⚠️ None, but untested                           |
| [Stereo audio](encoding/sad)         | .SAD           | ✅   | ✅    | ✅ None                                         |
| [Sprite](encoding/spr)               | .SPR           | ✅   | ❌    | ✅ None                                         |

## Tests

To run all tests:

```sh
go test -race ./...
```
