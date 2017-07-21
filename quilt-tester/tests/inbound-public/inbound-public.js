const {createDeployment} = require('@quilt/quilt');

let nginx = require('@quilt/nginx');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment({});
deployment.deploy(infrastructure);

for (let i = 0; i < infrastructure.nWorker; i++) {
    deployment.deploy(nginx.createService(80));
    deployment.deploy(nginx.createService(8000));
}
