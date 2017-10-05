const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');
const lobsters = require('@quilt/lobsters');

const deployment = new quilt.Deployment();
deployment.deploy(infrastructure);
lobsters.deploy(deployment, 'mysqlRootPassword');
