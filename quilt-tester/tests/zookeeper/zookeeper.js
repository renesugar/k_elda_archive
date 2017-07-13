const {createDeployment} = require('@quilt/quilt');
let zookeeper = require('@quilt/zookeeper');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment();
deployment.deploy(infrastructure);
deployment.deploy(new zookeeper.Zookeeper(infrastructure.nWorker*2));
