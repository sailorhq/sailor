source const.sh

echo "Installing sailor CLI version $VERSION"

cd ..
go build -o sailor -ldflags "-X main.Version=$VERSION" cmd/cli/main.go
sudo cp sailor /usr/local/bin
sudo chmod +x /usr/local/bin/sailor
rm sailor
