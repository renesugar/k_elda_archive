const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

console.log('This should show up in the terminal.');
console.warn('This too.');

const redContainer = new kelda.Container({ name: 'red', image: 'google/pause' });
redContainer.deploy(infra);
const blueContainer = new kelda.Container({ name: 'blue', image: 'google/pause' });
blueContainer.deploy(infra);
