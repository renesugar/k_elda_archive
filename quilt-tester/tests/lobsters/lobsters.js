const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');
const lobsters = require('@quilt/lobsters');

const deployment = quilt.createDeployment();
deployment.deploy(infrastructure);
lobsters.deploy(deployment, 'mysqlRootPassword');
