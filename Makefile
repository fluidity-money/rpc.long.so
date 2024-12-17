
rpc.long.so: $(shell find -type f -name '*.go')
	@go build

.PHONY: lambda

lambda: bootstrap.zip

bootstrap: rpc.long.so
	@cp rpc.long.so bootstrap

bootstrap.zip: bootstrap
	@zip bootstrap.zip bootstrap

clean:
	@rm bootstrap bootstrap.zip rpc.long.so
