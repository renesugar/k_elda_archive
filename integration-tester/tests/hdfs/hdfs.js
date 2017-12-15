const hadoop = require('@kelda/hadoop');
const infrastructure = require('../../config/infrastructure.js');

const hdfs = new hadoop.HDFS(infrastructure.nWorker - 1);

const infra = infrastructure.createTestInfrastructure();
hdfs.deploy(infra);
