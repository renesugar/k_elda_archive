# Compiles all of the docs in docs/build.
all: build/slate
	cd build/slate && \
	rsync -a ${CURDIR}/source/ source/ && \
	rsync -a ${CURDIR}/*md source/ && \
	bundle exec middleman build --clean && \
	rsync -a build/ ${CURDIR}/build/ && \
	echo "docs successfully compiled to HTML. To view docs, run:" && \
	echo "  open build/index.html"

build/slate:
	cd build && \
	git clone https://github.com/quilt/slate.git && \
	cd slate && \
	bundle install
