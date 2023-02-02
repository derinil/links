run: build
	@./bin/links

build:
	@go build -o bin/links .

lo:
	@docker-compose --file docker-compose.yaml up --build

dolo:
	@docker-compose --file docker-compose.yaml down

test:
	@go test ./...
