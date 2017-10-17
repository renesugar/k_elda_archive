const etcd = require('@kelda/etcd');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();
const etcdApp = new etcd.Etcd(infrastructure.nWorker * 2);
etcdApp.deploy(infra);
