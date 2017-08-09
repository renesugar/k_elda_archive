const quilt = require('@quilt/quilt');
let zookeeper = require('@quilt/zookeeper');
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);
deployment.deploy(new zookeeper.Zookeeper(infrastructure.nWorker*2));
