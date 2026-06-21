.PHONY: build
build:
	go build -o notes ./cmd/notes

.PHONY: install
install: build
	rm -f ~/go/bin/notes
	rm -f ~/.local/bin/nn
	cp notes ~/go/bin
	cp nn.sh ~/.local/bin/nn
