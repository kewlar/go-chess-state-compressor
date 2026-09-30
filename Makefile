# Variables
BIN_DIR  := bin
CMD_DIR  := cmd
COMMANDS := $(wildcard $(CMD_DIR)/*)
BINARIES := $(patsubst $(CMD_DIR)/%,$(BIN_DIR)/%,$(COMMANDS))

# Go build flags (-s -w strips debugging symbols to reduce binary size)
LDFLAGS  := -s -w
GOFLAGS  := -trimpath -ldflags="$(LDFLAGS)"

.PHONY: all clean list

# Default target: build all binaries found in cmd/
all: $(BINARIES)

# Pattern rule: build bin/<name> from cmd/<name>
$(BIN_DIR)/%: $(CMD_DIR)/% $(wildcard $(CMD_DIR)/%/*.go) $(wildcard internal/*/*.go)
	@mkdir -p $(BIN_DIR)
	@echo "Building $@..."
	@go build $(GOFLAGS) -o $@ ./$<

# Clean build artifacts
clean:
	@echo "Cleaning $(BIN_DIR)..."
	@rm -rf $(BIN_DIR)

# List discovered targets
list:
	@echo "Commands found in $(CMD_DIR)/:"
	@for cmd in $(COMMANDS); do echo "  - $$(basename $$cmd)"; done

