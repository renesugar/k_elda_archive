const spark = require('@quilt/spark');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const sprk = new spark.Spark(1, 3);

sprk.deploy(infra);
