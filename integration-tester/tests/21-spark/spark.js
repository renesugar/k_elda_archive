const kelda = require('kelda');
const spark = require('@kelda/spark');
const infrastructure = require('../../config/infrastructure.js');

// Application
// sprk.exposeUIToPublic says that the the public internet should be able
// to connect to the Spark web interface.
const sprk = new spark.Spark(infrastructure.nWorker - 1)
  .exposeUIToPublic();

// Add environment variables to the Driver to allow it to read input data from S3.
// These secrets are installed in the 20-spark-setup test, so they should be
// available when this blueprint runs.
sprk.driver.env.AWS_ACCESS_KEY_ID = new kelda.Secret('awsAccessKeyID');
sprk.driver.env.AWS_SECRET_ACCESS_KEY = new kelda.Secret('awsSecretAccessKey');

const infra = infrastructure.createTestInfrastructure();
sprk.deploy(infra);
