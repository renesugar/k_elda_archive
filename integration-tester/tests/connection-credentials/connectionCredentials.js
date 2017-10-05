const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const deployment = new quilt.Deployment();
deployment.deploy(infrastructure);

const container = new quilt.Container('red', 'google/pause');
container.deploy(deployment);
