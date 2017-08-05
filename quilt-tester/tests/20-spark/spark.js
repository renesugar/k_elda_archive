const {createDeployment} = require('@quilt/quilt');
let spark = require('@quilt/spark');
let infrastructure = require('../../config/infrastructure.js');

// Application
// sprk.exposeUIToPublic says that the the public internet should be able
// to connect to the Spark web interface. sprk.job causes Spark to run that
// job when it boots.
const sprk = new spark.Spark(1, infrastructure.nWorker-1)
  .exposeUIToPublic()
  .job('run-example SparkPi');

let deployment = createDeployment({});
deployment.deploy(infrastructure);
deployment.deploy(sprk);
