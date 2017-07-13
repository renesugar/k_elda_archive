// Place a google/pause container on each worker machine.

const {
    Container,
    Service,
    LabelRule,
    createDeployment} = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment({});
deployment.deploy(infrastructure);

let containers = new Service('containers',
    new Container('google/pause').replicate(infrastructure.nWorker));
containers.place(new LabelRule(true, containers));

deployment.deploy(containers);
