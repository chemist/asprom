.PHONY: all build test clean

all: build test

build:
	mkdir -p tmp/opt/asprom
	go build -i -o tmp/opt/asprom/asprom
	fpm -s dir -t rpm -n asprom --directories /opt/asprom --rpm-init etc/asprom -v 1.0 --iteration 4  -p ./ -C tmp

clean:
	rm -Rf tmp
	rm ./*.rpm
test:
	go test
