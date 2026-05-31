default:
  @just --list

build:
  go build -o bin/sprintctl main.go

test:
  go test -v ./...

clean:
  rm -rf bin/

tidy:
  go mod tidy

fmt:
  go fmt ./...

run:
  go run main.go

install:
  go install 'github.com/morethancertified/sprintctl'
  echo "sprintctl installed"

uninstall:
  rm -f $(go env GOPATH)/bin/sprintctl
  echo "sprintctl uninstalled"
