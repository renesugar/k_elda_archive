const {
    Container,
    Service,
    createDeployment,
    publicInternet} = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment();
deployment.deploy(infrastructure);

let connected = new Service('connected',
    new Container('alpine', ['tail', '-f', '/dev/null'])
        .replicate(infrastructure.nWorker*2)
);
publicInternet.allowFrom(connected, 80);

let notConnected = new Service('not-connected',
    new Container('alpine', ['tail', '-f', '/dev/null'])
        .replicate(infrastructure.nWorker*2)
);

deployment.deploy([connected, notConnected]);
