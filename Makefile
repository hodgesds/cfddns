
build/cfddns: vendor
	go build -o build/cfddns ./cfddns

vendor:
	@glide install

clean:
	@-rm -rf vendor build
