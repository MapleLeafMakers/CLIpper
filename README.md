![CLIpper](assets/clipper.png)

## CLIpper

A text-based interface for [Klipper](https://www.klipper3d.org/) & [Moonraker](https://github.com/Arksine/moonraker)

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![GitHub Release](https://img.shields.io/github/v/release/MapleLeafMakers/CLIpper?label=Release)](https://github.com/MapleLeafMakers/CLIpper/releases/latest)

## ⚠️ Warning

CLIpper is still in very early development, and lacks *many* features.  It is not currently an acceptable replacement for other interfaces unless you speak fluent gcode.

## Installation

CLIpper is available for Linux, macOS and Windows platforms. Binaries are available on the [release](https://github.com/MapleLeafMakers/CLIpper/releases/latest) page.

Just extract, and put the `clipper` or `clipper.exe` wherever you want.

## Usage

run the `clipper` executable, passing a hostname, ip address, or full url to the moonraker websocket server.

```shell
./clipper 192.168.1.100
./clipper trident.local
./clipper ws://mainsailos.local/websocket

```

## Building From Source

### Install Build Tools

CLIpper is written in [Go](https://go.dev/), download and install the relevant build tools, instructions can be found [here](https://go.dev/doc/install)


### Clone The Repository

```shell
git clone git@github.com:MapleLeafMakers/CLIpper.git
```

### Install Dependencies 

If you setup golang correctly, this should do it:

```shell
go mod download
```

### Build The Executable

This can be as simple as:

```shell
go build
```
    
You can also include version and git commit information in the build using:
    
```shell
go build -v -ldflags="-X main.commit=$(git describe --always --long --dirty) -X main.version=v0.0.0"
```

© 2024 MapleLeafMakers
