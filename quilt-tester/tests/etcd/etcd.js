const quilt = require('@quilt/quilt');
let etcd = require('@quilt/etcd');
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);
deployment.deploy(new etcd.Etcd(infrastructure.nWorker*2));
