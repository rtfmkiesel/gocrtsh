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
git clone https://gitlab.com/rtfmkiesel/gocrtsh
cd gocrtsh
# to build binary in the current directory
go build -ldflags="-s -w" .
# to build & install binary into GOPATH/bin
go install .
```