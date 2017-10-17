const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const container = new kelda.Container('red', 'google/pause');
container.deploy(infra);
