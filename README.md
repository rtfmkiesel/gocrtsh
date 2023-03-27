# gocrtsh
A quick & minimal implementation of [crt.sh](https://crt.sh)'s JSON API in Golang. 

## Usage
```
cat domains.txt | gocrtsh (-w) (-r 3)

-w = Print found wildcard (*.domain.tld) certificates (default: false)
-r = How many runners (threads) to spawn (default: 1)
```

## Installation
```bash
go install gitlab.com/rtfmkiesel/gocrtsh@latest
```

## Build from source
```bash
# manual
git clone https://gitlab.com/rtfmkiesel/gocrtsh
cd gocrtsh
# to build binary in the current directory
go build -ldflags="-s -w" "cli/gocrtsh"
# to install binary into GOPATH/bin
go install "cli/gocrtsh"

# via makefile
git clone https://gitlab.com/rtfmkiesel/gocrtsh
cd gocrtsh
# current OS
make
# cross compile
make all
```