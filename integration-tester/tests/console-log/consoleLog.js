const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

console.log('This should show up in the terminal.');
console.warn('This too.');

const redContainer = new quilt.Container('red', 'google/pause');
redContainer.deploy(infra);
const blueContainer = new quilt.Container('blue', 'google/pause');
blueContainer.deploy(infra);
