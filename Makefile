.PHONY: lint test install build 
all: install

lint: 
	bash ./scripts/install_lint.sh
	${GOPATH}/bin/golangci-lint run

test:
	go test ./... 

install:
	cd cmd/ && go install && cd ..

build:
	cd cmd/ && go build -o onioncli && mv onioncli .. && cd address && go build -o onionaddress && mv onionaddress ../.. && cd ../..
