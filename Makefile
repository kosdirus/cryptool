.PHONY:
.SILENT:

build:
	go build -o ./.bin/app cmd/server/main.go

run: build
	./.bin/app

build-image:
	docker build -t cryptool:v0.1 .

start-container:
	docker run --name cryptool-cont -p 8080:8080 --env-file .env cryptool:v0.1