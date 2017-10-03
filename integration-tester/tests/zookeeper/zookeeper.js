const quilt = require('@quilt/quilt');
const zookeeper = require('@quilt/zookeeper');
const infrastructure = require('../../config/infrastructure.js');

const deployment = quilt.createDeployment();
deployment.deploy(infrastructure);
const zk = new zookeeper.Zookeeper(infrastructure.nWorker * 2);
zk.deploy(deployment);
