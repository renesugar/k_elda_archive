const zookeeper = require('@quilt/zookeeper');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();
const zk = new zookeeper.Zookeeper(infrastructure.nWorker * 2);
zk.deploy(infra);
