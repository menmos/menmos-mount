export CGO_ENABLED := "1"

build $GOOS="linux" $GOARCH="amd64":
    @mkdir -p build/
    go build -o build/menmos_mount-$GOOS-$GOARCH ./cmd

clean:
    rm -rf build

unit:
    go test ./...

integration +args="":
    go test -tags integration -c ./testing/integration/... -o build/integration
    @./build/integration {{args}}
