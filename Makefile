BIN=gocrtsh

make: clean
	mkdir -p ./builds
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "./builds/${BIN}_lin_amd64"
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o "./builds/${BIN}_lin_arm64"
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "./builds/${BIN}_macos_amd64"
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "./builds/${BIN}_macos_arm64"
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "./builds/${BIN}_win_amd64.exe"
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o "./builds/${BIN}_win_arm64.exe"

clean:
	rm -rf ./builds
	go mod tidy
	go clean -r