# Quilt Documentation

The Quilt documentation is located at [docs.quilt.io](http://docs.quilt.io).
This directory contains the Markdown files that are used to generate those
docs. We do not recommend directly browing the Markdown files, because the
links are designed to work properly for the HTML version of the page, and
will not work correctly in the Markdown version.

If you're a developer and you'd like to build the docs, you'll first need
to install a few other tools. The docs are built (by compiling the Markdown
files here into a single HTML file) using Quilt's
[Slate fork](https://github.com/quilt/slate). The README in that repository
includes detailed installation instructions. After completing the
installation described there, you can use the `docs` build target in the
Quilt root directory: 

```console
$ make docs
```


