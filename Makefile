.PHONY: build lint unittest go-lint yaml-lint coverage

build:
	go build -o app src/main.go

go-lint:
	golangci-lint run --timeout 5m ./src/...
	golint src/...

yaml-lint:
	find ./ -name '*.yaml' -exec yamllint {} +

lint: | go-lint yaml-lint
	
unittest:
	go test ./...

coverage:
	go test -covermode=count -coverprofile cov --tags unit ./...
	go tool cover -html=cov -o coverage.html

clean:
	$(RM) app
