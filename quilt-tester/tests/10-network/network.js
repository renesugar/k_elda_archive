const quilt = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

let c = new quilt.Container('alpine', ['tail', '-f', '/dev/null']);
let red = new quilt.Service('red', c.replicate(5));
let blue = new quilt.Service('blue', c.replicate(5));
let yellow = new quilt.Service('yellow', c.replicate(5));

blue.allowFrom(red, 80);
red.allowFrom(blue, 80);
yellow.allowFrom(red, 80);
yellow.allowFrom(blue, 80);

deployment.deploy(red);
deployment.deploy(blue);
deployment.deploy(yellow);
