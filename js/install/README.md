This repo contains a Node.js package for easily installing
[Kelda](http://kelda.io) using npm.

To pull the compiled binaries off GitHub, and install them:
```
$ npm install -g @kelda/install
```
Note that the `-g` flag is necessary to install the binaries globally (i.e. so
that `kelda` can be invoked from anywhere).

If installing as root, the `--unsafe-perm` flag is required:
```console
$ sudo npm install -g @kelda/install --unsafe-perm
```
