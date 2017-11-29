const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const connected = [];
for (let i = 0; i < infrastructure.nWorker * 2; i += 1) {
  connected.push(new kelda.Container('outbound', 'appropriate/curl', {
    command: ['tail', '-f', '/dev/null'],
  }));
}

kelda.publicInternet.allowFrom(connected, 80);

const notConnected = [];
for (let i = 0; i < infrastructure.nWorker * 2; i += 1) {
  notConnected.push(new kelda.Container('ignoreme', 'appropriate/curl', {
    command: ['tail', '-f', '/dev/null'],
  }));
}

connected.forEach(connectedContainer => connectedContainer.deploy(infra));
notConnected.forEach(notConnectedContainer => notConnectedContainer.deploy(infra));
