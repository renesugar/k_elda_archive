const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

// If we deploy nWorker+1 containers, at least one machine is guaranteed to run
// two containers, and thus be able to test intra-machine bandwidth.
const iperfs = [];
for (let i = 0; i < infrastructure.nWorker + 1; i += 1) {
  iperfs.push(new quilt.Container('iperf', 'networkstatic/iperf3', {
    command: ['-s'],
  }));
}
quilt.allow(iperfs, iperfs, 5201);
iperfs.forEach(container => container.deploy(deployment));
