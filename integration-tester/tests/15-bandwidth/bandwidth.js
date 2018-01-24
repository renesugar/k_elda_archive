const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

// If we deploy nWorker+1 containers, at least one machine is guaranteed to run
// two containers, and thus be able to test intra-machine bandwidth.
const iperfs = [];
for (let i = 0; i < infrastructure.nWorker + 1; i += 1) {
  iperfs.push(new kelda.Container('iperf', 'networkstatic/iperf3', {
    command: ['-s'],
  }));
}
kelda.allowTraffic(iperfs, iperfs, 5201);
iperfs.forEach(container => container.deploy(infra));
