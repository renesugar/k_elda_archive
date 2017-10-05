const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

const container = new quilt.Container('red', 'google/pause');
container.deploy(deployment);
