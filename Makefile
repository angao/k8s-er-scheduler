.PHONY: build
.PHONY: clean

build: build-local

build-local:
	# build
	go build -o kube-er-scheduler *.go

clean:
	# delete build file
	rm -rf ./kube-er-scheduler