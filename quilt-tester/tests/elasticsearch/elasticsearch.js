const quilt = require('@quilt/quilt');
let Elasticsearch = require('@quilt/elasticsearch').Elasticsearch;
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);
deployment.deploy(new Elasticsearch(infrastructure.nWorker).allowFromPublic());
