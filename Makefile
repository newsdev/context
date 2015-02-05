all: bin/context

release: bin/context
	aws s3 cp bin/context s3://newsdev-pub/bin/context

bin/context: bin
	docker build -t context-build $(CURDIR) && docker run --rm -v $(CURDIR)/bin:/opt/bin context-build cp /go/bin/app /opt/bin/context

bin:
	mkdir -p bin

clean:
	rm -rf bin
	docker rmi context-build || true
