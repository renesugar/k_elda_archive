const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

console.log('This should show up in the terminal.');
console.warn('This too.');

const redContainer = new quilt.Container('red', 'google/pause');
redContainer.deploy(deployment);
const blueContainer = new quilt.Container('blue', 'google/pause');
blueContainer.deploy(deployment);
