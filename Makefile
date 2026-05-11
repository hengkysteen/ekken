.PHONY: build build-all ui clean tag

BINARY_NAME=ekken
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
LDFLAGS := -ldflags="-s -w -X 'ekken/internal/config.mode=production'"
VERSION := $(shell grep -E 'var buildVersion =' internal/config/config.go | awk -F'"' '{print $$2}')

clean:
	rm -rf dist/ ui/dist/

ui:
	cd ui && npm ci && npm run build

build: clean ui
	mkdir -p dist
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -tags production $(LDFLAGS) -o dist/$(BINARY_NAME) .

build-all: clean ui
	mkdir -p dist
	GOOS=windows GOARCH=amd64 go build -tags production $(LDFLAGS) -o dist/$(BINARY_NAME)-windows.exe .
	GOOS=linux   GOARCH=amd64 go build -tags production $(LDFLAGS) -o dist/$(BINARY_NAME)-linux .
	GOOS=darwin  GOARCH=amd64 go build -tags production $(LDFLAGS) -o dist/$(BINARY_NAME)-mac-intel .
	GOOS=darwin  GOARCH=arm64 go build -tags production $(LDFLAGS) -o dist/$(BINARY_NAME)-mac-silicon .

tag:
	@echo "Reading version from internal/config/config.go..."
	@echo "Current version: $(VERSION)"
	@if git show-ref --tags --verify --quiet "refs/tags/$(VERSION)"; then \
		echo "❌ Error: Tag $(VERSION) already exists!"; \
		exit 1; \
	else \
		git tag $(VERSION); \
		echo "✅ Tag $(VERSION) created successfully!"; \
		echo "👉 Next, run this command to push to remote:"; \
		echo "   git push origin $(VERSION)"; \
	fi
