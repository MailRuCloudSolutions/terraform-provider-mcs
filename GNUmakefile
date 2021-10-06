TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=mcs

default: build

build: fmtcheck
	go install

build_darwin: fmtcheck
	GOOS=darwin CGO_ENABLED=0 go build -o terraform-provider-mcs_darwin

build_linux: fmtcheck
	GOOS=linux CGO_ENABLED=0 go build -o terraform-provider-mcs_linux

build_windows: fmtcheck
	GOOS=windows CGO_ENABLED=0 go build -o terraform-provider-mcs_windows

test: fmtcheck
	go test $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -cover -timeout=30s -parallel=4

testmock_k8saas: fmtcheck
	TF_ACC=1 TF_ACC_MOCK_MCS=1 go test $(TEST) -run=TestMockAcc $(TESTARGS) -timeout 120m

testacc_k8saas: fmtcheck
	TF_ACC=1 go test -run=TestAccKubernetes $(TEST) $(TESTARGS) -timeout 120m

testacc_dbaas: fmtcheck
	TF_ACC=1 go test -run=TestAccDatabase $(TEST) -v $(TESTARGS) -timeout 120m

testacc_bs: fmtcheck
	TF_ACC=1 go test -run=TestAccBlockStorage -v $(TESTARGS) -timeout 120m

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1
	golangci-lint run ./...

.PHONY: build test testacc vet fmt fmtcheck errcheck test-compile website website-test lint

