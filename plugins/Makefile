# M3M Plugins Makefile
# Build all plugins or individual ones

GOBUILD = go build

# Find all plugin directories (those with go.mod)
PLUGIN_DIRS := $(shell find . -maxdepth 2 -name 'go.mod' -exec dirname {} \; | grep -v '^\./$$')

.PHONY: all clean list tidy

# Build all plugins
all:
	@echo "Building all plugins..."
	@for dir in $(PLUGIN_DIRS); do \
		name=$$(basename $$dir); \
		echo "  Building $$name..."; \
		(cd $$dir && $(GOBUILD) -buildmode=plugin -o ../$$name.so); \
	done
	@echo "Done! Built plugins:"
	@ls -la *.so 2>/dev/null || echo "  No .so files found"

# Build specific plugin
# Usage: make telegram
%:
	@if [ -d "./$@" ]; then \
		echo "Building $@..."; \
		cd ./$@ && $(GOBUILD) -buildmode=plugin -o ../$@.so; \
		echo "Built $@.so"; \
	else \
		echo "Plugin '$@' not found"; \
		exit 1; \
	fi

# Clean all .so files
clean:
	@echo "Cleaning plugin binaries..."
	@rm -f *.so
	@echo "Done!"

# List available plugins
list:
	@echo "Available plugins:"
	@for dir in $(PLUGIN_DIRS); do \
		name=$$(basename $$dir); \
		echo "  - $$name"; \
	done

# Tidy all plugin modules
tidy:
	@echo "Tidying plugin modules..."
	@for dir in $(PLUGIN_DIRS); do \
		name=$$(basename $$dir); \
		echo "  Tidying $$name..."; \
		(cd $$dir && go mod tidy); \
	done
	@echo "Done!"

help:
	@echo "M3M Plugins Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make all        - Build all plugins"
	@echo "  make <name>     - Build specific plugin (e.g., make telegram)"
	@echo "  make clean      - Remove all .so files"
	@echo "  make list       - List available plugins"
	@echo "  make tidy       - Run go mod tidy for all plugins"
	@echo "  make help       - Show this help"
