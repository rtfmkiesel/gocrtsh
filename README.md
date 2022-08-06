# gocrtsh
A quick & minimal implementation of [crt.sh](https://crt.sh)'s JSON API in GO. 

## Installation
```bash
go install gitlab.com/lu-ka/gocrtsh@latest
```

## Build from source
```bash
git clone https://gitlab.com/lu-ka/gocrtsh
cd gocrtsh
go build

# or via makefile
make linux || make windows || make all
```

## Usage
```
cat domains.txt | gocrtsh
```

## copycat
This is not the first nor only (GO) client for [crt.sh](https://crt.sh). I just wanted to do my own minimalistic implementation with learning by doing in mind. 

## License
This code is released under the [MIT License](https://gitlab.com/lu-ka/gocrtsh/blob/main/LICENSE).

## Legal
This code is provided for educational use only. If you engage in any illegal activity the author does not take any responsibility for it. By using this code, you agree with these terms.
