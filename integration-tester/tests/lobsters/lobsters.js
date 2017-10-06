const infrastructure = require('../../config/infrastructure.js');
const lobsters = require('@quilt/lobsters');

const infra = infrastructure.createTestInfrastructure();
lobsters.deploy(infra, 'mysqlRootPassword');
