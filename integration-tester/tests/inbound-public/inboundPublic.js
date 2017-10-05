const quilt = require('@quilt/quilt');

const nginx = require('@quilt/nginx');
const infrastructure = require('../../config/infrastructure.js');

const deployment = new quilt.Deployment();
deployment.deploy(infrastructure);

for (let i = 0; i < infrastructure.nWorker; i += 1) {
  nginx.createContainer(80).deploy(deployment);
  nginx.createContainer(8000).deploy(deployment);
}
