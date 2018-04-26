DIR=$(shell pwd)

up:
	echo $(DIR)
	cd $(DIR)/.. && GOOS=linux GOARCH=amd64 go build -v -o ./example/simple ./example/simple.go
	cd $(DIR) && docker-compose up --build

down:
	docker-compose down && docker-compose rm