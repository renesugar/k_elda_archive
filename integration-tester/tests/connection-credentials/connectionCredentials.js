const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const container = new quilt.Container('red', 'google/pause');
container.deploy(infra);
