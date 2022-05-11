TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=hashicorp.com
NAMESPACE=dev
NAME=multy
BINARY=terraform-provider-${NAME}
VERSION=0.0.1
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
INSTALL_PATH=~/.local/share/terraform/plugins/localhost/providers/hcs/0.0.1/linux_$(GOARCH)
BUILD_ALL_PATH=${PWD}/bin

ifeq ($(GOOS), "darwin")
	INSTALL_PATH=~/Library/Application\ Support/io.terraform/plugins/localhost/providers/hcs/0.0.1/darwin_$(GOARCH)
endif
ifeq ($(GOOS), "windows")
	INSTALL_PATH=%APPDATA%/HashiCorp/Terraform/plugins/localhost/providers/hcs/0.0.1/windows_$(GOARCH)
endif

default: install

build:
	go build -o ${BINARY}

release:
	goreleaser release --rm-dist --snapshot --skip-publish  --skip-sign

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${GOOS}_${GOARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${GOOS}_${GOARCH}

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

docs: FORCE
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

FORCE: ;