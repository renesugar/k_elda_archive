# Compiles all of the docs in docs/build.
all:
	rm -rf build && \
	mkdir build && \
	cd build && \
	git clone https://github.com/quilt/slate.git && \
	cd slate && \
	cp ${CURDIR}/*md source/ && \
	bundle install && \
	bundle exec middleman build --clean && \
	mv build/* ${CURDIR}/build/ && \
	echo "docs successfully compiled to HTML. To view docs, run:" && \
	echo "  open build/index.html"
