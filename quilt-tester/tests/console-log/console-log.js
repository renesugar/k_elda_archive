const {Container, Service, createDeployment} = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment({});
deployment.deploy(infrastructure);

console.log('This should show up in the terminal.');
console.warn('This too.');

deployment.deploy(new Service('red', [new Container('google/pause')]));
deployment.deploy(new Service('blue', [new Container('google/pause')]));
