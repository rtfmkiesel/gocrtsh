# gocrtsh
A quick & minimal implementation of [crt.sh](https://crt.sh)'s JSON API in Golang. 

## Usage
```
cat domains.txt | gocrtsh [OPTIONS]

Options:
    -r Print only resolvable domains          (default: false)
    -w Print found wildcard certificates      (default: false)
    -t How many threads to spawn              (default: 1)
    -s Do not print errors                    (default: false)
    -h Prints this text
```

## Installation
```bash
go install github.com/rtfmkiesel/gocrtsh@latest
```

## Build from source
```bash
git clone https://github.com/rtfmkiesel/gocrtsh
cd gocrtsh
# to build binary in the current directory
go build -ldflags="-s -w" .
# to build & install binary into GOPATH/bin
go install .
```