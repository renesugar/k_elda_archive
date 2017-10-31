const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const numContainers = Math.round(infrastructure.nWorker / 2);
for (let i = 0; i < numContainers; i += 1) {
  (new kelda.Container('testContainer', 'alpine', {
    command: ['tail', '-f', '/dev/null'],
  })).deploy(infra);
}
