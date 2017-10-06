const infrastructure = require('../../config/infrastructure.js');
const Redis = require('@quilt/redis');

const infra = infrastructure.createTestInfrastructure();

const redis = new Redis(3, 'password');
redis.deploy(infra);
