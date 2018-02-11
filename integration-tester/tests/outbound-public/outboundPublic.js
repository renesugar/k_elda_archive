const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const connected = [];
for (let i = 0; i < infrastructure.nWorker * 2; i += 1) {
  connected.push(new kelda.Container({
    name: 'outbound',
    image: 'appropriate/curl',
    command: ['tail', '-f', '/dev/null'],
  }));
}

kelda.allowTraffic(connected, kelda.publicInternet, 80);

const notConnected = [];
for (let i = 0; i < infrastructure.nWorker * 2; i += 1) {
  notConnected.push(new kelda.Container({
    name: 'ignoreme',
    image: 'appropriate/curl',
    command: ['tail', '-f', '/dev/null'],
  }));
}

connected.forEach(connectedContainer => connectedContainer.deploy(infra));
notConnected.forEach(notConnectedContainer => notConnectedContainer.deploy(infra));
