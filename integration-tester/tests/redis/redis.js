const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');
const Redis = require('@quilt/redis');

const deployment = new quilt.Deployment();
deployment.deploy(infrastructure);

const redis = new Redis(3, 'password');
redis.deploy(deployment);
