const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

(new kelda.Container({
  name: 'privileged',
  image: 'ubuntu',
  command: ['tail', '-f', '/dev/null'],
  privileged: true,
})).deploy(infra);

(new kelda.Container({
  name: 'not-privileged',
  image: 'ubuntu',
  command: ['tail', '-f', '/dev/null'],
})).deploy(infra);
