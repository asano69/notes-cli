
.PHONY: build
build:
	go build -o notes ./cmd/notes

.PHONY: install
install:
	rm ~/go/bin/notes
	cp notes ~/go/bin/notes
