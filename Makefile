GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test -v
GOCLEAN=$(GOCMD) clean
GOGET = $(GOCMD) get -d
GOFORMAT = $(GOCMD) fmt
GOINSTALL = $(GOCMD) install

REMOTE_PACKS= \
	fyne.io/fyne/v2 \
	github.com/disintegration/imaging \
	github.com/sirupsen/logrus \
	github.com/sqweek/dialog \
	gotest.tools/v3 \

REMOTE_BIN = \
	fyne.io/fyne/v2/cmd/fyne \


SRCS=main.go
BINNAME=Carve
BINDIR=./bin
BIN=$(BINDIR)/$(BINNAME)
ICON=./app_icon.png
APP=GoCarver

build:
# build: fetch
# test: fetch
deploy: build
run: deploy
all: fetch test build

build:
	$(GOBUILD) -o $(BIN) $(SRCS)
test:
	$(GOTEST) ./...
fetch:
	$(GOGET) $(REMOTE_PACKS)
	$(GOCMD) mod vendor
clean:
	$(GOCLEAN)
	rm -f $(BIN)
format:
	$(GOFORMAT) $(SRCS)
update:
	$(GOGET) -u $(REMOTE_PACKS)
	$(GOINSTALL) $(REMOTE_BIN)
deploy:
	fyne package -os darwin -exe $(BIN)  -release
	rm -rf ./deploy/$(APP).app > /dev/null 2>&1
	mv -f $(APP).app ./deploy
	mkdir -p ./deploy
run:
	deploy/$(APP).app/Contents/MacOS/$(BINNAME)