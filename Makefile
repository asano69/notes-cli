.PHONY: build
build:
	go build -o notes ./cmd/notes

.PHONY: install
install: build
	rm -f ~/go/bin/notes
	rm -f ~/.local/bin/notes-editor
	cp notes ~/go/bin
	cp notes-editor ~/.local/bin/notes-editor
