source const.sh

cd ..

rm -f release/*

build_sailor_binaries() {
    GOOS=windows GOARCH=amd64 go build -o $1-windows -ldflags "-X main.Version=$VERSION" cmd/$2/main.go
    GOOS=darwin GOARCH=arm64 go build -o $1-mac-arm64 -ldflags "-X main.Version=$VERSION" cmd/$2/main.go
    GOOS=darwin GOARCH=amd64 go build -o $1-mac-amd64 -ldflags "-X main.Version=$VERSION" cmd/$2/main.go
    GOOS=linux GOARCH=amd64 go build -o $1-linux-amd64 -ldflags "-X main.Version=$VERSION" cmd/$2/main.go

    zip release/$1-$VERSION-windows-amd64.zip $1-windows
    zip release/$1-$VERSION-mac-arm64.zip $1-mac-arm64
    zip release/$1-$VERSION-mac-amd64.zip $1-mac-amd64
    zip release/$1-$VERSION-linux-amd64.zip $1-linux-amd64


    rm $1-windows
    rm $1-mac-arm64
    rm $1-mac-amd64
    rm $1-linux-amd64
}

build_sailor_binaries "sailor" "cli"
build_sailor_binaries "sailor-core" "sailor"