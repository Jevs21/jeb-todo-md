BINARY := jeb-todo-md
CMD    := ./cmd/jeb-todo-md

.PHONY: build test run clean

build:
	go build -o $(BINARY) $(CMD)

test:
	go test -v ./...

run: build
	JEB_TODO_FILE=$${JEB_TODO_FILE:?set JEB_TODO_FILE} ./$(BINARY)

clean:
	rm -f $(BINARY)
