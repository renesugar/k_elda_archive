const nginx = require('@quilt/nginx');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

for (let i = 0; i < infrastructure.nWorker; i += 1) {
  nginx.createContainer(80).deploy(infra);
  nginx.createContainer(8000).deploy(infra);
}
