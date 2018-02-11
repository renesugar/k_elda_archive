const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const image = new kelda.Image({
  name: 'network-image',
  dockerfile: 'FROM ubuntu\n' +
  'RUN apt-get update\n' +
  'RUN apt-get install -y iputils-ping hping3\n',
});
const command = ['tail', '-f', '/dev/null'];

const redContainers = [];
for (let i = 0; i < infrastructure.nWorker; i += 1) {
  redContainers.push(new kelda.Container({ name: 'red', image, command }));
}

const blueContainers = [];
for (let i = 0; i < infrastructure.nWorker; i += 1) {
  blueContainers.push(new kelda.Container({ name: 'blue', image, command }));
}

const yellowContainers = [];
for (let i = 0; i < infrastructure.nWorker; i += 1) {
  yellowContainers.push(new kelda.Container({ name: 'yellow', image, command }));
}

kelda.allowTraffic(redContainers, blueContainers, 80);
kelda.allowTraffic(blueContainers, redContainers, 80);
kelda.allowTraffic(redContainers, yellowContainers, 80);
kelda.allowTraffic(blueContainers, yellowContainers, 80);

redContainers.forEach(container => container.deploy(infra));
yellowContainers.forEach(container => container.deploy(infra));
blueContainers.forEach(container => container.deploy(infra));
