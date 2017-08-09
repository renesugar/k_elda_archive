const {createDeployment} = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment();
deployment.deploy(infrastructure);
