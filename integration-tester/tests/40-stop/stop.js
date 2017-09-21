const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

const containers = new quilt.Container('foo', 'google/pause')
  .replicate(infrastructure.nWorker);
deployment.deploy(containers);
