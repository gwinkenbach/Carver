# Copyright 2020 The Chromium OS Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test -v
GOCLEAN=$(GOCMD) clean
GOGET = $(GOCMD) get
GOFORMAT = $(GOCMD) fmt

REMOTE_PACKS= \
	github.com/therecipe/qt/cmd/... \
	gotest.tools/v3

SRCS=main.go
BIN=./bin/Carve

all: build
build: fetch
test: fetch
deploy: build
run: deploy

build:
	$(GOBUILD) -o $(BIN) $(SRCS)
test:
	$(GOTEST) ./...
fetch:
	$(GOGET) $(REMOTE_PACKS)
clean:
	$(GOCLEAN)
	rm -f $(BIN)
format:
	$(GOFORMAT) $(SRCS)
update:
	$(GOGET) -u $(REMOTE_PACKS)
deploy:
	qtdeploy -qt_version 5.13.2 build desktop
run:
	deploy/darwin/GoCarver.app/Contents/MacOS/GoCarver