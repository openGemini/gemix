[![LICENSE](https://img.shields.io/github/license/openGemini/gemix.svg)](https://github.com/openGemini/gemix/blob/master/LICENSE)
[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/openGemini/gemix)](https://goreportcard.com/badge/github.com/openGemini/gemix)
[![Coverage Status](https://codecov.io/gh/openGemini/gemix/branch/master/graph/badge.svg)](https://codecov.io/gh/openGemini/gemix/)

# What is gemix

`gemix` is a tool to download and install [openGemini](https://github.com/openGemini/openGemini) components.

## Documentation

- [English]()
- [简体中文](https://docs.opengemini.org/zh/guide/reference/gemix_manual.html)

## Installation

```sh
go install github.com/openGemini/gemix@latest
```

## Quick start

### Run playground (TODO)

```sh
gemix playground
```

### Deploy cluster

```sh
gemix cluster insall opengemini_demo v1.1.1 ./topology.yaml --skip-create-user -p
```

Tips:

> If there are network issues with the automatic download of the installation package, 
> please place the installation package under `~/.gemix/storage/cluster/packages/`

## Usage

After installing `gemix`, you can use it to install binaries of openGemini components and create clusters.

## Contributing to gemix

Contributions of code, tests, docs, and bug reports are welcome! To get started, take a look at our [open issues](https://github.com/openGemini/gemix/issues).

## License

[![LICENSE](https://img.shields.io/github/license/openGemini/gemix.svg)](https://github.com/openGemini/gemix/blob/master/LICENSE)
