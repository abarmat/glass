prepare-dev:
	docker-compose up -d
	migrate -source file://migrations -database "postgres://glass:glass@localhost:5432/glass?sslmode=disable" up

build:
	go build -o glass -v

run-dev: prepare-dev run

run:
	go run main.go
