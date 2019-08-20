# Transactor-UI

[![Build Status](https://travis-ci.org/katena-chain/transactor-ui.svg?branch=v0.0.1)](https://travis-ci.org/katena-chain/transactor-ui)

## Requirements

This project uses [golang](https://golang.org/).

**To compile on Ubuntu/Debian, you may need to install the ```libgl1-mesa-dev``` and ```xorg-dev``` packages, and have GCC installed.**

### Tested versions

In order to run the project properly, some tools are required:

- [golang](https://golang.org/) (Tested: v1.12.6)

## Installation

Install go-bindata:

```bash
go get -u github.com/go-bindata/go-bindata
go install github.com/go-bindata/go-bindata
```

Generate assets:

```bash
go generate gui/main.go
```

Build binary:
```bash
go build -o build/transactor-ui cmd/main.go
```

## Using the tool

Run binary :
```bash
./build/transactor-ui
```

## Releases

You'll find the release binaries under the ``build`` folder. Run it like the above using the corresponding path.

### Cross-compiling

To cross-compile for Windows and MacOS, build the Dockerfile and use the corresponding image.

For example, from the root, run : 
```bash
docker run -v ${PWD}:/app transchain/golang-crosscompile:v1.0.0 goreleaser --rm-dist --skip-publish --snapshot
```
