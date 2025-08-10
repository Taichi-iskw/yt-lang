.PHONY: test unit lint fmt e2e

unit:
	go test ./... -count=1 -race -shuffle=on
test:
	unit
lint:
	go vet ./...
fmt:
	gofmt -s -w .
e2e:
	go test -tags=integration ./... -count=1