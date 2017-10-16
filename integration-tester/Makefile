# Get the names of the test suites by getting the relative path to all
# *_test.go files, getting the parent directory, and stripping the leading
# path. For example, tests/100-logs/logs_test.go gets converted to
# tests/100-logs, and finally 100-logs.
suites_with_test_files := $(shell basename $(dir $(wildcard tests/*/*_test.go)))
test_binaries := $(foreach name, $(suites_with_test_files), tests/$(name)/$(name).test)

.PHONY: tests
tests: $(test_binaries)

%.test:
	cd $(@D) && go test -v -c .
