const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

for (let workerIndex = 0; workerIndex < infrastructure.nWorker; workerIndex += 1) {
  const image = new quilt.Image(`test-custom-image${workerIndex}`,
    'FROM alpine\n' +
    `RUN echo ${workerIndex} > /dockerfile-id\n` +
    'RUN echo $(cat /dev/urandom | tr -dc \'a-zA-Z0-9\' | ' +
    'fold -w 32 | head -n 1) > /image-id');
  for (let containerIndex = 0; containerIndex < 2; containerIndex += 1) {
    const container = new quilt.Container(
      'bar', image, { command: ['tail', '-f', '/dev/null'] });
    container.deploy(infra);
  }
}
