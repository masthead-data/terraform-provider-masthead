default: fmt lint install generate

clean:
	rm -rf examples/provider/.terraform
	rm -rf examples/provider/.terraform.lock.hcl

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	make clean
	make generate
	go install .
	go test -v -cover -timeout=120s -parallel=10 ./...
	terraform -chdir=examples/provider init
	export TF_LOG="DEBUG"
	terraform -chdir=examples/provider apply -auto-approve

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: fmt lint test testacc build install generate
