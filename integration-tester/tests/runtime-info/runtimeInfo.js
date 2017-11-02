const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');
const crypto = require('crypto');

const infra = infrastructure.createTestInfrastructure();

const indexPath = '/usr/share/nginx/html/index.html';

const webContainer = new kelda.Container('web', 'nginx', {
  env: {
    myPubIP: kelda.hostIP,
  },
  filepathToContent: {
    [indexPath]: crypto.randomBytes(20).toString('hex'),
  },
});

webContainer.allowFrom(kelda.publicInternet, 80);
webContainer.deploy(infra);
