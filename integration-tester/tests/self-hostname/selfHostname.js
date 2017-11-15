const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const indexPath = '/usr/share/nginx/html/index.html';
const infra = infrastructure.createTestInfrastructure();

const webContainers = [];
for (let i = 0; i < infrastructure.nWorker; i += 1) {
  const webContainer = new kelda.Container('web', 'nginx');

  // Make the container return its hostname when queried. The test relies on
  // this to check that its query routed to the correct container.
  webContainer.filepathToContent[indexPath] = webContainer.getHostname();
  webContainers.push(webContainer);
  webContainer.deploy(infra);
}

const fetcherContainer = new kelda.Container('fetcher', 'alpine', {
  command: ['tail', '-f', '/dev/null'],
});
kelda.allow(fetcherContainer, webContainers, 80);
fetcherContainer.deploy(infra);
