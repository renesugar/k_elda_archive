const {
    Container,
    PortRange,
    Service,
    createDeployment} = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment({});
deployment.deploy(infrastructure);

let nWorker = 1;
let red = new Service('red', new Container('google/pause').replicate(nWorker));
let blue = new Service(
    'blue', new Container('google/pause').replicate(3 * nWorker));

let ports = new PortRange(1024, 65535);
blue.allowFrom(red, ports);
red.allowFrom(blue, ports);

deployment.deploy(red);
deployment.deploy(blue);
