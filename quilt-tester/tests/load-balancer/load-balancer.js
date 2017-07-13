const {createDeployment, Container, Service} = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment({});
deployment.deploy(infrastructure);

let containers = [];
for (let i = 0; i < 4; i++) {
  containers.push(new Container('nginx:1.10').withFiles({
    '/usr/share/nginx/html/index.html':
        'I am container number ' + i.toString() + '\n',
  }));
}

let fetcher = new Service('fetcher',
    [new Container('alpine', ['tail', '-f', '/dev/null'])]);
let loadBalanced = new Service('loadBalanced', containers);
loadBalanced.allowFrom(fetcher, 80);

deployment.deploy([fetcher, loadBalanced]);
