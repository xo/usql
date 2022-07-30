PACKAGE_NAME := github.com/xo/usql

PWD := $(shell pwd)

.PHONY: prepare
prepare:
	docker build -f .github/Dockerfile -t osxcross-usql .

.PHONY: release
release:
	docker run --rm --privileged \
		-e CGO_ENABLED=1 \
		-e GITHUB_TOKEN \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(PWD):/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		osxcross-usql \
		--rm-dist --skip-validate
