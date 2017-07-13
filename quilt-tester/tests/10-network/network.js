const {createDeployment, Service, Container} = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment({});
deployment.deploy(infrastructure);

let c = new Container('alpine', ['tail', '-f', '/dev/null']);
let red = new Service('red', c.replicate(5));
let blue = new Service('blue', c.replicate(5));
let yellow = new Service('yellow', c.replicate(5));

blue.allowFrom(red, 80);
red.allowFrom(blue, 80);
yellow.allowFrom(red, 80);
yellow.allowFrom(blue, 80);

deployment.deploy(red);
deployment.deploy(blue);
deployment.deploy(yellow);
