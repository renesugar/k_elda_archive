const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const deployment = new quilt.Deployment();
deployment.deploy(infrastructure);

const containers = [];
for (let i = 0; i < 4; i += 1) {
  containers.push(new quilt.Container('web', 'nginx:1.10').withFiles({
    '/usr/share/nginx/html/index.html':
        `I am container number ${i.toString()}\n`,
  }));
}
containers.forEach(container => container.deploy(deployment));

const fetcher = new quilt.Container('fetcher', 'alpine', {
  command: ['tail', '-f', '/dev/null'],
});
const loadBalanced = new quilt.LoadBalancer('loadBalanced', containers);
loadBalanced.allowFrom(fetcher, 80);

fetcher.deploy(deployment);
loadBalanced.deploy(deployment);
