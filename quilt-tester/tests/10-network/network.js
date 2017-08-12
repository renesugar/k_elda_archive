const quilt = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

let c = new quilt.Container('alpine', {
    command: ['tail', '-f', '/dev/null'],
});

let red = new quilt.Service('red', c.replicate(5));
setHostnames(red.containers, 'red');

let blue = new quilt.Service('blue', c.replicate(5));
setHostnames(blue.containers, 'blue');

let yellow = new quilt.Service('yellow', c.replicate(5));
setHostnames(blue.containers, 'blue');

blue.allowFrom(red, 80);
red.allowFrom(blue, 80);
yellow.allowFrom(red, 80);
yellow.allowFrom(blue, 80);

deployment.deploy(red);
deployment.deploy(blue);
deployment.deploy(yellow);

/**
 * setHostnames sets the hostnames for `containers` to be a unique hostname
 * prefixed by `hostname`.
 *
 * @param {quilt.Container[]} containers
 * @param {string} hostname
 */
function setHostnames(containers, hostname) {
    containers.forEach((c) => {
        c.setHostname(hostname);
    });
}
