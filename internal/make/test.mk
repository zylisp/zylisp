test: test-core test-common test-parser test-lexer test-reader

test-core:
	@echo "\nRunning 'go test' for core ...\n" && \
	cd ./core && \
	GO111MODULE=on go test -v

test-common:
	@echo "\nRunning 'go test' for common ...\n" && \
	cd ./common && \
	GO111MODULE=on go test -v

test-parser:
	@echo "\nRunning 'go test' for parser ...\n" && \
	cd ./core/parser && \
	GO111MODULE=on go test -v

test-lexer:
	@echo "\nRunning 'go test' for lexer ...\n" && \
	cd ./core/lexer && \
	GO111MODULE=on go test -v

test-reader:
	@echo "\nRunning 'go test' for reader ...\n" && \
	cd ./core/reader && \
	GO111MODULE=on go test -v

# Example usage:
#  $ TEST=Backup make run-lexer-test
run-lexer-test:
	@echo "\nRunning 'go test' for lexer test $(TEST)...\n" && \
	cd ./core/lexer && \
	GO111MODULE=on go test -v -run LexerSuite -testify.m $(TEST)

gogen-examples:
	@echo "\nGenerating .go files for examples ...\n"
	$(ZY) -cli -go -dir ./bin/examples examples/*.zsp

ast-examples:
	@echo "\nGenerating AST files for examples ...\n"
	$(ZY) -cli -ast -dir ./examples examples/*.zsp

bin/examples/%: bin/examples/%.go
	GO111MODULE=on go build -o $@ $<

build-examples: gogen-examples ast-examples
	@echo "\nCompiling example go files ...\n"
	@$(MAKE) $(basename $(wildcard ./bin/examples/*.go))

clean-examples:
	@echo "Removing generated example files ..."
	@rm -rf ./bin/examples ./examples/*.ast ./examples/*.go

test-cli:
	./tests/test-zylisp-cli.sh

test-examples:
	./tests/test-compiled-examples.sh

test-zyc:
	./tests/test-zyc.sh

bench-inner-outer:
	go test -v -run=^$ -bench=. ./play/func_call_benchmark_test.go