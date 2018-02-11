const kelda = require('kelda');
const infrastructure = require('../../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const numContainers = parseInt(process.argv[2], 10);
const name = process.argv[3];

for (let i = 0; i < numContainers; i += 1) {
  const container = new kelda.Container({ name, image: 'google/pause' });
  container.deploy(infra);
}
