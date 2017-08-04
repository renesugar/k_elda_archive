const {createDeployment} = require('@quilt/quilt');
let Elasticsearch = require('@quilt/elasticsearch').Elasticsearch;
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment({});
deployment.deploy(infrastructure);
deployment.deploy(new Elasticsearch(infrastructure.nWorker).allowFromPublic());
