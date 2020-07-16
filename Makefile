run:
	go run -race cmd/main.go

build:
	go build cmd/main.go

test:
	go test -v ./... -cover

docker-build:
	docker build -t lucasalmeron/backtaboo .

docker-push:
	docker push lucasalmeron/backtaboo

compose-build:
	docker-compose build

docker-up: 
	docker-compose up

docker-down:
	docker-compose down