const spark = require('@kelda/spark');
const infrastructure = require('../../config/infrastructure.js');

// Application
// sprk.exposeUIToPublic says that the the public internet should be able
// to connect to the Spark web interface. sprk.job causes Spark to run that
// job when it boots.
const sprk = new spark.Spark(1, infrastructure.nWorker - 1)
  .exposeUIToPublic();

const infra = infrastructure.createTestInfrastructure();
sprk.deploy(infra);
