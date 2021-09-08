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

pull_latest TARGET_DIR="menmos_bin":
    #!/usr/bin/env bash
    tag=$(curl --silent "https://api.github.com/repos/menmos/menmos/releases/latest" | grep -Po '"tag_name": "\K.*?(?=")')
    mkdir -p {{TARGET_DIR}}
    curl -L -o {{TARGET_DIR}}/menmosd https://github.com/menmos/menmos/releases/download/$tag/menmosd-linux-amd64
    curl -L -o {{TARGET_DIR}}/amphora https://github.com/menmos/menmos/releases/download/$tag/amphora-linux-amd64
