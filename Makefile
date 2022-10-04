
export PACKAGE := $(abspath .)
export OUTDIR := $(abspath ./bin)

export GOBIN := ${OUTDIR}:${GOPATH}/bin:${GOBIN}
export PATH := ${PATH}:${GOBIN}
export TMPDIR := /tmp

export TARGET := breath
export GIT_SSH_COMMAND='ssh -o ControlMaster=no'

.PHONY: all build run mock

all: build run

build:
	mkdir -p ${OUTDIR}
	go get -d .
	go build -o ${OUTDIR}/${TARGET}

run:
	${OUTDIR}/${TARGET}
