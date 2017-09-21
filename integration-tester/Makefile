SHELL := /bin/bash

.PHONY: tests
tests:
	for suite in tests/* ; do \
		pushd "$$suite" > /dev/null && \
		go test -v -c . ; \
		popd > /dev/null ; \
	done
