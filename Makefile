NAME      = devcloud
BUILD_DIR = $(CURDIR)/build

.PHONY: all clean
all: clean www

clean:
	rm -rf $(BUILD_DIR)

www: www-linux-amd64 www-darwin-amd64 www-darwin-arm64

www-linux-amd64 www-darwin-amd64 www-darwin-arm64:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$$(echo $@ | sed 's/^www-//;s/-.*//') GOARCH=$$(echo $@ | sed 's/^www-//;s/.*-//') \
		go build -o $(BUILD_DIR)/$(NAME)-$@ ./cmd/www
