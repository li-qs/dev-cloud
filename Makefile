NAME       := devcloud
BUILD_DIR  := $(CURDIR)/build
CMD_DIR    := .
GOFLAGS    ?=
# 交叉编译目标：<os>/<arch>
PLATFORMS  := linux/amd64 darwin/amd64 darwin/arm64

.PHONY: all build cross run vet fmt tidy test clean

all: clean cross

# 编译当前平台
build:
	@mkdir -p $(BUILD_DIR)
	go build $(GOFLAGS) -o $(BUILD_DIR)/$(NAME) $(CMD_DIR)

# 交叉编译多平台（纯静态，CGO 关闭）
cross:
	@mkdir -p $(BUILD_DIR)
	@for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; \
		echo "building $(NAME)-$$os-$$arch"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			go build $(GOFLAGS) -o $(BUILD_DIR)/$(NAME)-$$os-$$arch $(CMD_DIR) || exit 1; \
	done

# 本地开发运行（ENV=development 时自动建表）
run:
	ENV=development go run $(CMD_DIR) -config ./config.yaml

vet:
	go vet ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)
