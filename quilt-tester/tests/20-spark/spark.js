const {createDeployment, publicInternet, enough} = require('@quilt/quilt');
let spark = require('@quilt/spark');
let infrastructure = require('../../config/infrastructure.js');

// Application
// sprk.exclusive enforces that no two Spark containers should be on the
// same node. sprk.public says that the containers should be allowed to talk
// on the public internet. sprk.job causes Spark to run that job when it
// boots.
let sprk = new spark.Spark(1, infrastructure.nWorker-1)
    .exclusive()
    .public()
    .job('run-example SparkPi');

let deployment = createDeployment({});
deployment.deploy(infrastructure);
deployment.deploy(sprk);

deployment.assert(publicInternet.canReach(sprk.masters), true);
deployment.assert(enough, true);
