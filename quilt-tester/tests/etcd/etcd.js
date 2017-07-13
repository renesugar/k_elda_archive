const {createDeployment} = require('@quilt/quilt');
let etcd = require('@quilt/etcd');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment({});
deployment.deploy(infrastructure);
deployment.deploy(new etcd.Etcd(infrastructure.nWorker*2));
