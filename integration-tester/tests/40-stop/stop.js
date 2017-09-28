const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

const containers = [];
for (let i = 0; i < infrastructure.nWorker; i += 1) {
  containers.push(new quilt.Container('foo', 'google/pause'));
}
deployment.deploy(containers);
