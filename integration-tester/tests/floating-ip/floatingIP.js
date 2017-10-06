const nginx = require('@quilt/nginx');
const infrastructure = require('../../config/infrastructure.js');

// The integration tests always run in the default region for each provider,
// so we only need to consider the provider when assigning floating IPs.
const providerToFloatingIp = {
  Amazon: '13.57.99.49', // us-west-1
  Google: '104.196.11.66', // us-east1-b
  DigitalOcean: '138.68.203.188', // sfo1
};

const infra = infrastructure.createTestInfrastructure();

// Find a worker machine to which we'll assign a floating IP.
let aWorker;
for (let i = 0; i < infra.machines.length; i += 1) {
  const m = infra.machines[i];
  if (m.role === 'Worker') {
    aWorker = m;
    break;
  }
}
if (aWorker === undefined) {
  throw new Error('Failed to find any worker machines');
}

const floatingIp = providerToFloatingIp[aWorker.provider];
if (floatingIp === undefined) {
  throw new Error(`No floating IP for provider ${aWorker.provider}`);
}

// Because `aWorker` references the machine within `infra`, assigning
// to the floatingIp here automatically updates the infrastructure.
aWorker.floatingIp = floatingIp;

const nginxContainer = nginx.createContainer(80);
nginxContainer.placeOn(aWorker);
nginxContainer.deploy(infra);
