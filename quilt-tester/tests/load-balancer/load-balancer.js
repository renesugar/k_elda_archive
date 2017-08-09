const quilt = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

let containers = [];
for (let i = 0; i < 4; i++) {
  containers.push(new quilt.Container('nginx:1.10').withFiles({
    '/usr/share/nginx/html/index.html':
        'I am container number ' + i.toString() + '\n',
  }));
}

let fetcher = new quilt.Service('fetcher',
    [new quilt.Container('alpine', ['tail', '-f', '/dev/null'])]);
let loadBalanced = new quilt.Service('loadBalanced', containers);
loadBalanced.allowFrom(fetcher, 80);

deployment.deploy([fetcher, loadBalanced]);
