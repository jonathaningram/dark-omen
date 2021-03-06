# Warhammer: Dark Omen Go library

[![GoDoc](https://godoc.org/github.com/jonathaningram/dark-omen?status.svg)](http://godoc.org/github.com/jonathaningram/dark-omen)
[![Build status](https://github.com/jonathaningram/dark-omen/workflows/Go/badge.svg?branch=master)](https://github.com/jonathaningram/dark-omen/actions)
[![Report card](https://goreportcard.com/badge/github.com/jonathaningram/dark-omen)](https://goreportcard.com/report/github.com/jonathaningram/dark-omen)
[![Sourcegraph](https://sourcegraph.com/github.com/jonathaningram/dark-omen/-/badge.svg)](https://sourcegraph.com/github.com/jonathaningram/dark-omen?badge)

This library is for developers interested in building tools for working with Dark Omen's assets.

This library does not ship with any Dark Omen assets. You must have a legally purchased copy of Dark Omen in order to get the benefits of this library.

**Note:** this library is neither developed by nor endorsed by Electronic Arts Inc.

## Table of contents

- [Installation](#installation)
- [Tests](#tests)

## Installation

Use `go get` to retrieve the library and add it to your project's Go module dependencies.

```shell
go get github.com/jonathaningram/dark-omen
```

To update the library use `go get -u` to retrieve the latest version of the library.

```shell
go get -u github.com/jonathaningram/dark-omen
```

## Tests

To run all tests:

```sh
go test -race ./...
```
