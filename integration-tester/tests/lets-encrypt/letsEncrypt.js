const kelda = require('kelda');
const hap = require('@kelda/haproxy_auto_https');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const indexPath = '/usr/share/nginx/html/index.html';

/**
 * Returns a new Container whose index file contains the given content.
 * @param {string} content - The contents to put in the container's index file.
 * @returns {Container} - A container with given content in its index file.
 */
function containerWithContent(content) {
  return new kelda.Container('web', 'nginx', {
    filepathToContent: {
      [indexPath]: content,
    },
  });
}

const serviceA = [
  containerWithContent('a1'),
  containerWithContent('a2'),
];

const serviceB = [
  containerWithContent('b1'),
  containerWithContent('b2'),
  containerWithContent('b3'),
];

const workersWithFloatingIps = infra.machines.filter(
  m => m.role === 'Worker' && m.floatingIp !== '');
if (workersWithFloatingIps.length !== 1) {
  throw new Error('There should be exactly one floating IP assigned to a Worker');
}
const floatingIp = workersWithFloatingIps[0].floatingIp;
const provider = workersWithFloatingIps[0].provider;

const email = 'dev@kelda.io';
const domainA = `ci-${provider}-a.kelda.io`;
const domainB = `ci-${provider}-b.kelda.io`;

const domainToContainers = {};
domainToContainers[domainA] = serviceA;
domainToContainers[domainB] = serviceB;

const proxy = hap.create(domainToContainers, email, { testing_cert: true });

proxy.allowFrom(kelda.publicInternet, hap.exposedHttpPort);
proxy.allowFrom(kelda.publicInternet, hap.exposedHttpsPort);

serviceA.forEach(container => container.deploy(infra));
serviceB.forEach(container => container.deploy(infra));
proxy.placeOn({ floatingIp });
proxy.deploy(infra);
