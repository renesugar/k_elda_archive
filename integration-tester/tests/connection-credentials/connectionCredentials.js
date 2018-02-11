const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const container = new kelda.Container({ name: 'red', image: 'google/pause' });
container.deploy(infra);
