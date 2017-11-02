const infrastructure = require('../../config/infrastructure.js');
const Redis = require('@kelda/redis');

const infra = infrastructure.createTestInfrastructure();

const redis = new Redis(infrastructure.nWorker, 'password');
redis.deploy(infra);
