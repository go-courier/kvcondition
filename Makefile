test:
	go test -v -race ./...

cover: test
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...
