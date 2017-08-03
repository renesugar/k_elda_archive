const {
    Container,
    Service,
    createDeployment} = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment({});
deployment.deploy(infrastructure);

let containers = new Service('containers',
    new Container('google/pause').replicate(infrastructure.nWorker));
deployment.deploy(containers);
