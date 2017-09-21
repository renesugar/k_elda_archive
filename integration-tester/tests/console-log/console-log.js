const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

console.log('This should show up in the terminal.');
console.warn('This too.');

deployment.deploy(new quilt.Container('red', 'google/pause'));
deployment.deploy(new quilt.Container('blue', 'google/pause'));
