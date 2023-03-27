BIN=gocrtsh

make: clean
	go build -ldflags="-s -w" -o "${BIN}" "cli/${BIN}"

all: clean
	mkdir -p ./builds
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "./builds/${BIN}_lin_amd64" "cli/${BIN}"
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o "./builds/${BIN}_lin_arm64" "cli/${BIN}"
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "./builds/${BIN}_macos_amd64" "cli/${BIN}"
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "./builds/${BIN}_macos_arm64" "cli/${BIN}"
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "./builds/${BIN}_win_amd64.exe" "cli/${BIN}"
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o "./builds/${BIN}_win_arm64.exe" "cli/${BIN}"

clean:
	rm -rf ./builds
	go mod tidy
	go clean -r