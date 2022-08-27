include .project/gomod-project.mk
export GO111MODULE=on
BUILD_FLAGS=
export GOPRIVATE=github.com/ableorg
export COVERAGE_EXCLUSIONS="vendor|tests|api/pb/gw|main.go|testsuite.go|mocks.go|.pb.go|.pb.gw.go"

.PHONY: *

.SILENT:

default: help

all: clean tools generate version test

#
# clean produced files
#
clean:
	go clean ./...
	rm -rf \
		${COVPATH} \
		${PROJ_BIN}

tools:
	go install golang.org/x/tools/cmd/stringer
	go install github.com/go-phorce/cov-report/cmd/cov-report
	go install golang.org/x/lint/golint
	go install github.com/princjef/gomarkdoc/cmd/gomarkdoc

version:
	echo "*** building version"
	gofmt -r '"GIT_VERSION" -> "$(GIT_VERSION)"' internal/version/current.template > internal/version/current.go

build:
	echo "*** Building email-sync"
	go build ${BUILD_FLAGS} -o ${PROJ_ROOT}/bin/email-sync ./cmd/email-sync
