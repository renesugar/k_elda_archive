const quilt = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

let containers = new quilt.Container('foo', 'google/pause')
  .replicate(infrastructure.nWorker);
deployment.deploy(containers);
