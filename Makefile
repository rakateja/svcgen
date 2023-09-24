run:
	go run main.go
	go fmt out/domains/**/*.go

runv2:
	go run *.go --out=rpc --target=../product

install:
	go build -o svcgen main.go
	mv svcgen /usr/local/bin/svcgen