const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const deployment = new quilt.Deployment();
deployment.deploy(infrastructure);

for (let i = 0; i < infrastructure.nWorker; i += 1) {
  const container = new quilt.Container('foo', 'google/pause');
  container.deploy(deployment);
}
