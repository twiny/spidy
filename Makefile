clean:
	@echo "Removing ~/.spidy/"
	@rm -rf ~/.spidy/

start:
	@./bin/spidy start --output ./results --config config/config.json csv --file ./testdata/list.csv --column url_column

init:
	@./bin/spidy init

build:
	@echo "Building spidy..."
	@go build -v -o ./bin/spidy cmd/spidy/main.go
	@echo "Build complete."

usage:
	htop -p $(pgrep -d',' -f ./spidy)

.PHONY: clean start init build usage