const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const images = [];
for (let imageIndex = 0; imageIndex < 8; imageIndex += 1) {
  const image = new kelda.Image(`test-custom-image${imageIndex}`,
    'FROM alpine\n' +
    `RUN echo ${imageIndex} > /dockerfile-id\n` +
    'RUN echo $(cat /dev/urandom | tr -dc \'a-zA-Z0-9\' | ' +
    'fold -w 32 | head -n 1) > /image-id\n' +
    'RUN sleep 15');

  images.push(image);
}


for (let workerIndex = 0; workerIndex < infrastructure.nWorker; workerIndex += 1) {
  for (let containerIndex = 0; containerIndex < images.length; containerIndex += 1) {
    const container = new kelda.Container(
      'bar', images[containerIndex], { command: ['tail', '-f', '/dev/null'] });
    container.deploy(infra);
  }
}
