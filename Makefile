build:
	docker build --no-cache -t yestool/deploy-tar:v1 .

build-cli:
	go build -o ./client/bin/deploy-tar ./client/main.go