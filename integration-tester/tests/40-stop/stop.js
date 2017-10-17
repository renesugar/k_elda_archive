const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

for (let i = 0; i < infrastructure.nWorker; i += 1) {
  const container = new kelda.Container('foo', 'google/pause');
  container.deploy(infra);
}
