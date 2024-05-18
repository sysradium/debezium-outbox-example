GO ?= CGO_ENABLED=0 go
GOOS ?= linux

build: $(foreach dir,$(wildcard cmd/*),build-$(notdir $(dir)))

build-%:
	GOOS=$(GOOS) $(GO) build -o ./bin/$* ./cmd/$*

run:
	go run ./cmd/nats
