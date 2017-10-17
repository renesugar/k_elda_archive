const kelda = require('kelda');
const hap = require('@kelda/haproxy');
const infrastructure = require('../../config/infrastructure.js');

const indexPath = '/usr/share/nginx/html/index.html';

/**
 * Returns a new Container whose index file contains the given content.
 * @param {string} content - The contents to put in the container's index file.
 * @returns {Container} - A container with given content in its index file.
 */
function containerWithContent(content) {
  return new kelda.Container('web', 'nginx', {
    filepathToContent: {
      [indexPath]: content,
    },
  });
}

const serviceA = [
  containerWithContent('a1'),
  containerWithContent('a2'),
];

const serviceB = [
  containerWithContent('b1'),
  containerWithContent('b2'),
  containerWithContent('b3'),
];

const proxy = hap.withURLrouting({
  'serviceB.com': serviceB,
  'serviceA.com': serviceA,
});

proxy.allowFrom(kelda.publicInternet, 80);

const inf = infrastructure.createTestInfrastructure();
serviceA.forEach(container => container.deploy(inf));
serviceB.forEach(container => container.deploy(inf));
proxy.deploy(inf);
