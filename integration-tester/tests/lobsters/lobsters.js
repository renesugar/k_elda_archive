const infrastructure = require('../../config/infrastructure.js');
const lobsters = require('@kelda/lobsters');

const infra = infrastructure.createTestInfrastructure();
lobsters.deploy(infra, 'mysqlRootPassword');
