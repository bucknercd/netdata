BINARY := netdata
CMD_DIR := ./cmd/netdata

.PHONY: build run clean rebuild

build:
	go build -o $(BINARY) $(CMD_DIR)

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)

rebuild: clean build

