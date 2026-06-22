.PHONY: build
build:
	go build -o notes ./cmd/notes

.PHONY: install
install: build
	rm -f ~/go/bin/notes
	rm -f ~/go/bin/notes-edit
	cp notes ~/go/bin
	cp notes-edit ~/go/bin
