run:
	@air --build.cmd "go build -o ./tmp/main ./cmd/paycue" --build.bin "./tmp/main"

build:
	go build -o ./bin/paycue ./cmd/paycue
	go build -o ./bin/paycue-cli ./cmd/paycue-cli

fmt:
	gofmt -w .

vet:
	go vet ./...
