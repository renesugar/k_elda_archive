const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const containers = [];
for (let i = 0; i < infrastructure.nWorker; i += 1) {
  containers.push(new kelda.Container('web', 'nginx:1.10', {
    filepathToContent: {
      '/usr/share/nginx/html/index.html':
      `I am container number ${i.toString()}\n`,
    } }));
}
containers.forEach(container => container.deploy(infra));

const fetcher = new kelda.Container('fetcher', 'alpine', {
  command: ['tail', '-f', '/dev/null'],
});
const loadBalanced = new kelda.LoadBalancer('loadBalanced', containers);
loadBalanced.allowFrom(fetcher, 80);

fetcher.deploy(infra);
loadBalanced.deploy(infra);
