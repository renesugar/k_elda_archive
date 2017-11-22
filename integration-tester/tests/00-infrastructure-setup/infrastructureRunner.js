const infrastructure = require('../../config/infrastructure.js');
const kelda = require('kelda');

const infra = infrastructure.createTestInfrastructure();

// Because each container listens for connections from the public internet they
// will need to be placed on separate machines. Therefore, the integration-tester
// will block until all the machines are ready, since every machine has to be
// utilized in order for all the containers to boot.
for (let i = 0; i < infrastructure.nWorker; i += 1) {
  const container = new kelda.Container('ignoreme', 'google/pause');
  container.allowFrom(kelda.publicInternet, 80);
  container.deploy(infra);
}
