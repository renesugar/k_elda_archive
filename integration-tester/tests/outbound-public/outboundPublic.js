const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const deployment = new quilt.Deployment();
deployment.deploy(infrastructure);

const connected = [];
for (let i = 0; i < infrastructure.nWorker * 2; i += 1) {
  connected.push(new quilt.Container('outbound', 'alpine', {
    command: ['tail', '-f', '/dev/null'],
  }));
}

quilt.publicInternet.allowFrom(connected, 80);

const notConnected = [];
for (let i = 0; i < infrastructure.nWorker * 2; i += 1) {
  notConnected.push(new quilt.Container('ignoreme', 'alpine', {
    command: ['tail', '-f', '/dev/null'],
  }));
}

connected.forEach(connectedContainer => connectedContainer.deploy(deployment));
notConnected.forEach(notConnectedContainer => notConnectedContainer.deploy(deployment));
