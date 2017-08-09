const quilt = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

console.log('This should show up in the terminal.');
console.warn('This too.');

deployment.deploy(new quilt.Service('red',
  [new quilt.Container('google/pause')]));
deployment.deploy(new quilt.Service('blue',
  [new quilt.Container('google/pause')]));
