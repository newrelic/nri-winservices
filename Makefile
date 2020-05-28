INTEGRATION     := winservices
BINARY_NAME      = nri-$(INTEGRATION)
WORKDIR         := $(shell pwd)
TARGET           = target
TARGET_DIR       = $(WORKDIR)/$(TARGET)
GO_FILES        := ./src/
GOOS             = GOOS=windows
GO               = $(GOOS) go

all: build

build: clean compile test

clean: compile-deps
	@echo "=== $(INTEGRATION) === [ clean ]: removing binaries and coverage file..."
	@rm -rfv $(TARGET_DIR) coverage.xml $(TARGET)

compile-deps:
	@echo "=== $(INTEGRATION) === [ compile-deps ]: installing build dependencies..."
	@$(GO) get -v -d -t ./...

bin/$(BINARY_NAME):
	@echo "=== $(INTEGRATION) === [ compile ]: building $(BINARY_NAME)..."
	@$(GO) build -v -o $(TARGET_DIR)/$(BINARY_NAME).exe $(GO_FILES)

compile: compile-deps bin/$(BINARY_NAME)

.PHONY: all build clean  compile-deps compile
