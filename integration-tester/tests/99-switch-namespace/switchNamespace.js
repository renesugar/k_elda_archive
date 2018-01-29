const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const numContainers = Math.round(infrastructure.nWorker);
for (let i = 0; i < numContainers; i += 1) {
  const c = new kelda.Container('test-container', 'nginx', {
    command: ['sh', '-c',
      'date > /usr/share/nginx/html/index.html && nginx -g "daemon off;"'],
  });
  kelda.allowTraffic(kelda.publicInternet, c, 80);
  c.deploy(infra);
}
