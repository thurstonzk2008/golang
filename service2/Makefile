export tag=v1.0
root:
	export ROOT=github.com/thurstonzk2008/golang

build:
	echo "building httpserver binary"
	mkdir -p bin/amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/amd64 .

release: build
	echo "building httpserver container"
	docker build -t thurstonzk2008/service2:${tag} .

push: release
	echo "pushing thurstonzk2008/service2"
	docker push thurstonzk2008/service2:v1.0
