const quilt = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

const image = 'alpine';
const command = ['tail', '-f', '/dev/null'];

let red = new quilt.Service('red-lb',
  new quilt.Container('red', image, {command}).replicate(5)
);
let blue = new quilt.Service('blue-lb',
  new quilt.Container('blue', image, {command}).replicate(5)
);
let yellow = new quilt.Service('yellow-lb',
  new quilt.Container('yellow', image, {command}).replicate(5)
);

blue.allowFrom(red, 80);
red.allowFrom(blue, 80);
yellow.allowFrom(red, 80);
yellow.allowFrom(blue, 80);

deployment.deploy(red);
deployment.deploy(blue);
deployment.deploy(yellow);
