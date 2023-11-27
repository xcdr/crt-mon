AUTHOR			= Adam Kubica <xcdr@kaizen-step.com>
BUILD_VERSION	= 0.2.3-rc
BUILD_BRANCH	= $(shell git rev-parse --abbrev-ref HEAD)
BUILD_DATE		= $(shell date +%Y%m%d%H%M)

BUILD_DIR		= build

LDFLAGS			+= -X 'main.author=${AUTHOR}'
LDFLAGS			+= -X 'main.version=${BUILD_VERSION}'
LDFLAGS			+= -X 'main.build=${BUILD_DATE}.${BUILD_BRANCH}'

all: build

show-version:
	@echo ${BUILD_BRANCH}-${BUILD_VERSION}

prepare:
	mkdir -p ${BUILD_DIR}/etc

build_linux-amd64: prepare
	mkdir -p ${BUILD_DIR}/linux-amd64/bin

	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/linux-amd64/bin/crt-mon cmd/crt-mon/main.go

	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/linux-amd64/bin/crt-check cmd/crt-check/main.go

build_linux-arm64: prepare
	mkdir -p ${BUILD_DIR}/linux-arm64/bin

	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/linux-arm64/bin/crt-mon cmd/crt-mon/main.go

	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/linux-arm64/bin/crt-check cmd/crt-check/main.go

build_darwin-arm64: prepare
	mkdir -p ${BUILD_DIR}/darwin-arm64/bin

	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
		go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/darwin-arm64/bin/crt-mon cmd/crt-mon/main.go

	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
		go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/darwin-arm64/bin/crt-check cmd/crt-check/main.go

build: build_linux-amd64 build_linux-arm64 build_darwin-arm64
	cp example.yml ${BUILD_DIR}/etc

artifacts: build
	cp install/systemd/crt-mon.service ${BUILD_DIR}/linux-amd64/
	cd ${BUILD_DIR} && tar czf linux-amd64.tar.gz -C linux-amd64 bin ../etc

	cp install/systemd/crt-mon.service ${BUILD_DIR}/linux-arm64/
	cd ${BUILD_DIR} && tar czf linux-arm64.tar.gz -C linux-arm64 bin ../etc

	cd ${BUILD_DIR} && tar czf darwin-arm64.tar.gz -C darwin-arm64 bin ../etc

clean:
	rm -rf ${BUILD_DIR}
