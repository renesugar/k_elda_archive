const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const indexPath = '/usr/share/nginx/html/index.html';

const webContainer = new kelda.Container('web', 'nginx', {
  env: {
    myPubIP: kelda.hostIP,
  },
  // Serve a random hash so that the test can ensure it really reached this
  // container. Although it would be nice for the hash to be randomly generated
  // at runtime, the blueprint would then produce two different containers when
  // run twice. This would confuse the integration-tester because it would think
  // that the container from the first run never booted.
  filepathToContent: {
    [indexPath]: 'bb48133faa41ecb860ffce0f4ee09a82',
  },
});

webContainer.allowFrom(kelda.publicInternet, 80);
webContainer.deploy(infra);
