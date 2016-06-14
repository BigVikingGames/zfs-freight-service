BIN_NAME=zfs-freight

test: clean-test
	dd if=/dev/zero of=$(CURDIR)/test.img bs=1 count=0 seek=2G
	zpool create test $(CURDIR)/test.img
	GO_ENV=test FREIGHT_ZPOOL=test go test -v -cover
	zpool destroy test
	rm $(CURDIR)/test.img

run:
	sudo -E go run main.go driver.go

clean-test:
	zpool destroy test || true
	rm $(CURDIR)/test.img || true

binary: clean-bin
	go build -o bin/$(BIN_NAME) -v

clean-bin:
	rm -Rf bin
