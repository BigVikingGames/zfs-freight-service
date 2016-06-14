BIN_NAME=zfs-freight
GO=$(shell which go)

test: clean-test
	dd if=/dev/zero of=$(CURDIR)/test.img bs=1 count=0 seek=2G
	zpool create test $(CURDIR)/test.img
	GO_ENV=test FREIGHT_ZPOOL=test $(GO) test -v -cover
	zpool destroy test
	rm $(CURDIR)/test.img

run:
	sudo -E $(GO) run main.go driver.go

clean-test:
	zpool destroy test || true
	rm $(CURDIR)/test.img || true

binary: clean-bin
	$(GO) build -o bin/$(BIN_NAME) -v

clean-bin:
	rm -Rf bin
