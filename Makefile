# https://confluence.lazada.com/display/RE/Requirements+to+build+process+for+GO+components

PWD = $(shell pwd)
GOPATH = ~/go

GOVER:=1.10
INSTALL_PATH:=/tmp/go${GOVER}
GO:=/usr/local/go/bin/go # default linux path

$(info trying to touch $(GO)…)
ifeq ($(wildcard $(GO)),)
	GO=go # default alias
endif

$(info checking version of $(GO)…)
GOVER_CURR:=$(shell $(GO) version | cut -d" " -f3 | sed 's/go//')
ifeq ($(findstring $(GOVER),$(GOVER_CURR)),)
$(info -- found GO $(GOVER_CURR), but required version is $(GOVER). Use custom build...)
	GO=$(INSTALL_PATH)/go/bin/go
	export GOROOT:=${INSTALL_PATH}/go
	export GLIDE_GO_EXECUTABLE=${GO}
else
$(info -- found GO $(GOVER), use current version...)
endif

GLIDE_VERSION:=v0.12.3
GLIDE_PATH:=$(GOPATH)/src/github.com/Masterminds/glide
GLIDE_BIN:=glide
GLIDE_INSTALLED := $(shell command -v glide 2> /dev/null)

ifdef GLIDE_INSTALLED
ifeq ($(findstring $(GLIDE_VERSION),$(shell $(GLIDE_BIN) --version | cut -d" " -f3 | sed 's/-dev//')),)
	GLIDE_BIN=$(GLIDE_PATH)/glide
endif
endif

ifndef GLIDE_INSTALLED
	GLIDE_BIN=$(GLIDE_PATH)/glide
endif

BIN?=$(lastword $(subst :, ,$(GOPATH)))/bin/motify_core_api
DATE:=$(shell date -u "+%Y-%m-%d %H:%M:%S")
VER:=$(shell git branch|grep '*'| cut -f2 -d' ')
GITHASH_SHORT:=$(shell git rev-parse --short HEAD)
DIRTY:=$(shell [ -n "$(shell git status --porcelain)" ] && echo ~dirty)
GITDESCRIBE := $(shell git describe --tags --long)
LDFLAGS=-X 'main.AppVersion=$(VER)($(GITHASH_SHORT))$(DIRTY)' -X 'main.GoVersion=$(GOVER)' -X 'main.BuildDate=$(DATE)' -X 'main.GitDescribe=$(GITDESCRIBE)'
LDFLAGS_STATIC= $(LDFLAGS) -extldflags '-static'
PROJECT_PATH:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
WTI_TOKEN:=qBjA4uSupkKSv8x5-1wjOQ
BINDATA_BIN:=$(GOPATH)/bin/go-bindata
BINDATA_ASSETFS_BIN:=$(GOPATH)/bin/go-bindata-assetfs
OS_NAME=$(shell uname -s | perl -ne 'print lc')
TEST_COVERAGE_OUTPUT:=acc.out


#vendor and glide aware settings
TEST_ARGS?=$$($(GLIDE_BIN) nv)

$(GOPATH)/bin/glide: |
	@echo "Installing glide ..."
	${GO} get -u github.com/Masterminds/glide

deps:                  ##install project dependencies (and install glide if you have no glide)
deps: get-glide
	$(info #Install dependencies...)
	$(GLIDE_BIN) install --force

get-glide: ${GO}
ifdef GLIDE_INSTALLED
ifeq ($(findstring $(GLIDE_VERSION),$(shell glide --version | cut -d" " -f3 | sed 's/-dev//')),)
	$(info found $(shell glide --version) but $(GLIDE_VERSION) required. Use custom glide build...)
	$(MAKE) glide-get-and-compile
else
$(info found applicable $(GLIDE_VERSION) in $(GLIDE_BIN)...)
endif
endif
ifndef GLIDE_INSTALLED
	$(info #Installing glide version $(GLIDE_VERSION)...)
	$(MAKE) glide-get-and-compile
endif

glide-get-and-compile:
	rm -rf $(GLIDE_PATH) ;\
	mkdir -p $(GLIDE_PATH) && cd $(GLIDE_PATH) ;\
	git clone https://github.com/Masterminds/glide.git . ;\
	rm -rf ./vendor ;\
	git fetch --all && git checkout -f $(GLIDE_VERSION) ;\
	make clean && make build ;\


build:                 ##building binary after install dependencies (you don't need to run "make deps" before "make build")
build:	deps fast-build

build-ci: deps fast-build

build-static: deps get-bindata download-translations ${GO}
	${BINDATA_ASSETFS_BIN} -pkg=motify_core_api translations/...
	${GO} build -ldflags "$(LDFLAGS_STATIC)" -o $(BIN) ./web/core_api/main.go

fast-build: 
	${GO} build -ldflags "$(LDFLAGS)" -o $(BIN) -i ./web/core_api/main.go

get-bindata: ${GO}
ifeq ($(wildcard $(BINDATA_BIN)),)
	$(info #Installing go-bindata ...)
	${GO} get -u github.com/jteeuwen/go-bindata/...
endif
ifeq ($(wildcard $(BINDATA_ASSETFS_BIN)),)
	$(info #Installing go-bindata-assetfs ...)
	${GO} get -u github.com/elazarl/go-bindata-assetfs/...
endif

get-ej: ${GO}
ifeq ($(wildcard ${GOPATH}/bin/easyjson),)
	$(info #Install easyjson...)
	${GO} get -u github.com/rilinor/easyjson/...
endif

ej: get-ej
	grep -rl --include='*.go' --exclude-dir='vendor' 'easyjson:json' . | xargs -P 4 -I {} bash -c 'echo "easyjson {}"; ${GOPATH}/bin/easyjson {}'

download-translations: ##download translations from webtranslateit
	$(info #Cleaning previous translations, downloading and preparing new one in "$(PROJECT_PATH)/translations"...)
ifeq (, $(shell which python))
	$(error "No python found, you need to install python")
else
	python $(PROJECT_PATH)/deploy/download-translations.py $(PROJECT_PATH) $(WTI_TOKEN)
endif

fast-test:
	$(info #Running tests...)
	${GO} test $(${GO} list ${PWD}/... | grep -v /vendor/ | grep -v /godep_libs/)

# make test
test:                  ##run unit tests
test: get-glide get-bindata ${GO}
	$(info #Running tests...)
	${BINDATA_ASSETFS_BIN} -pkg=motify_core_api README.md
	${GO} test $(TEST_ARGS)

test-verbose: get-glide get-bindata ${GO}
	${BINDATA_ASSETFS_BIN} -pkg=motify_core_api README.md
	${GO} test -v $(TEST_ARGS)

# make test-coverage
test-coverage:         ##run go coverage by packages
test-coverage: get-glide get-bindata ${GO}
	$(info #Running tests with coverage. Output in ${TEST_COVERAGE_OUTPUT}...)
	${BINDATA_ASSETFS_BIN} -pkg=motify_core_api README.md
	@# it's workaround for go coverage. It canot serve few packages by one run
	@echo "mode: set" > ${TEST_COVERAGE_OUTPUT}; \
	fail=0 ;\

	find . -maxdepth 10 ! -path '*/.*' ! -path './vendor/*' -type d | uniq | while read dir;  do \
		if ls $$dir/*_test.go >/dev/null 2>/dev/null; then \
			printf "== Directory $$dir == \n" ;\
			${GO} test $$dir -v -race -coverprofile=profile.out ;\
			if [ $$? != 0 ] ; then \
				fail=1 ;\
			fi; \
			if [ -f profile.out ] ;	then \
			  cat profile.out | grep -v "mode: set"| grep -v "mode: atomic" >> ${TEST_COVERAGE_OUTPUT};\
			  rm profile.out ;\
			fi ;\
		fi ;\
	done;\
	exit $$fail

# Go
/tmp/go${GOVER}:
	mkdir /tmp/go${GOVER}

/tmp/go${GOVER}/go${GOVER}.${OS_NAME}-amd64.tar.gz: /tmp/go${GOVER}
	curl -so /tmp/go${GOVER}/go${GOVER}.${OS_NAME}-amd64.tar.gz https://storage.googleapis.com/golang/go${GOVER}.${OS_NAME}-amd64.tar.gz

/usr/local/go: /tmp/go${GOVER}/go${GOVER}.${OS_NAME}-amd64.tar.gz
	cd /tmp/go${GOVER}/ && tar -xzf go${GOVER}.${OS_NAME}-amd64.tar.gz && touch go/bin/go

/tmp/go${GOVER}/go/bin/go: /tmp/go${GOVER}/go${GOVER}.${OS_NAME}-amd64.tar.gz
	cd /tmp/go${GOVER}/ && tar -xzf go${GOVER}.${OS_NAME}-amd64.tar.gz && touch go/bin/go

go:

LINT_EXCLUDE:=\
	--exclude='.*_easyjson\.go'\
	--exclude=".*_test\.go"
	--exclude="bindata_assetfs\.go"\
	--exclude="api\/core\/utils\/convertto\.go:\d+:2:warning: result assigned and not used \(ineffassign\)"

LINT:=gometalinter $(LINT_EXCLUDE)\
	--vendor\
	--deadline=60s\
	--cyclo-over=200\
	--min-occurrences=8\
	--line-length=278\
	--disable-all\
	--enable=vet\
	--enable=vetshadow\
	--enable=gosimple\
	--enable=staticcheck\
	--enable=ineffassign\
	--enable=gocyclo\
	--enable=lll\
	--enable=goconst\
	--enable=deadcode\
	./...

HINT:=gohint -config="deploy/go_hint_config.json"
GOHINT_ARGS:=$$(find . ! -path "./vendor/*" -iname '*.go')

lint-dep:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

ci-lint:              ## run gohint and gometalinter and make reports
ci-lint: lint-dep
	$(HINT) -reporter=checkstyle $(GOHINT_ARGS) > checkstyle-result1.xml || true
	$(LINT) --concurrency=1 --checkstyle > checkstyle-result2.xml || true

help:                  ##show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

SEDFLAG:=
ifeq ($(OS_NAME),darwin)
  SEDFLAG:=-i ''
endif

proto:
	protoc -I. --go_out=plugins=grpc:. ./api/ext_services/suggest_api/suggestions/suggestions.proto
	##replacing ",omitempty" in suggestions.pb.go file
	sed ${SEDFLAG} -e 's/,omitempty//g' ./api/ext_services/suggest_api/suggestions/suggestions.pb.go

.PHONY: fast-build download-translations build build-ci deps test proto save
