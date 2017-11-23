const nginx = require('@kelda/nginx');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const workersWithFloatingIps = infra.machines.filter(
  m => m.role === 'Worker' && m.floatingIp !== '');
if (workersWithFloatingIps.length !== 1) {
  throw new Error('There should be exactly one floating IP assigned to a Worker');
}

const nginxContainer = nginx.createContainer(80);
nginxContainer.placeOn({ floatingIp: workersWithFloatingIps[0].floatingIp });
nginxContainer.deploy(infra);
