const quilt = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

const image = 'alpine';
const command = ['tail', '-f', '/dev/null'];

let red = new quilt.Container('red', image, {command}).replicate(5);
let blue = new quilt.Container('blue', image, {command}).replicate(5);
let yellow = new quilt.Container('yellow', image, {command}).replicate(5);

quilt.allow(red, blue, 80);
quilt.allow(blue, red, 80);
quilt.allow(red, yellow, 80);
quilt.allow(blue, yellow, 80);

const redService = new quilt.Service('red-lb', red);
redService.allowFrom(blue, 80);

deployment.deploy(red);
deployment.deploy(yellow);
deployment.deploy(blue);
deployment.deploy(redService);
