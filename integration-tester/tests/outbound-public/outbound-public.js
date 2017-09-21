const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

const connected = new quilt.Container('outbound', 'alpine', {
  command: ['tail', '-f', '/dev/null'],
}).replicate(infrastructure.nWorker * 2);
quilt.publicInternet.allowFrom(connected, 80);

const notConnected = new quilt.Container('ignoreme', 'alpine', {
  command: ['tail', '-f', '/dev/null'],
}).replicate(infrastructure.nWorker * 2);

deployment.deploy(connected);
deployment.deploy(notConnected);
