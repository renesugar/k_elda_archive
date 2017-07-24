const {
    Container,
    Service,
    createDeployment} = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment({});
deployment.deploy(infrastructure);

let red = new Service('red', [new Container('google/pause')]);
deployment.deploy(red);
