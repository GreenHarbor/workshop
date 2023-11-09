lint:
	@docker run -t --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.55.1 golangci-lint run -v

run:
	@docker compose up

test:
	@go get -v -d ./... && go test ./tests