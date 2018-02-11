const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

// myEnvSecret and myFileSecret are defined in the 70-secret-setup test.
(new kelda.Container({
  name: 'ignoreme',
  image: 'alpine',
  command: ['tail', '-f', '/dev/null'],
  env: {
    MY_SECRET: new kelda.Secret('myEnvSecret'),
    NOT_A_SECRET: 'plaintext',
  },
  filepathToContent: {
    '/notASecret': 'plaintext',
    '/mySecret': new kelda.Secret('myFileSecret'),
  },
})).deploy(infra);
