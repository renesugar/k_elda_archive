const {createDeployment} = require('@quilt/quilt');

let nginx = require('@quilt/nginx');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment({});
deployment.deploy(infrastructure);

for (let i = 0; i < infrastructure.nWorker; i++) {
     /* eslint new-cap: "warn"*/
    deployment.deploy(nginx.New(80));
    deployment.deploy(nginx.New(8000));
}
