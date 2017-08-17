const quilt = require('@quilt/quilt');

let nginx = require('@quilt/nginx');
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

for (let i = 0; i < infrastructure.nWorker; i++) {
    deployment.deploy(nginx.createContainer(80));
    deployment.deploy(nginx.createContainer(8000));
}
