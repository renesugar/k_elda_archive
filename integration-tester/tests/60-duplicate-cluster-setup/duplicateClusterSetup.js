const quilt = require('@quilt/quilt');
const spark = require('@quilt/spark');
const infrastructure = require('../../config/infrastructure.js');

const deployment = new quilt.Deployment();
deployment.deploy(infrastructure);

const sprk = new spark.Spark(1, 3);

sprk.deploy(deployment);
