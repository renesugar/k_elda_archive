const quilt = require('@quilt/quilt');
const Elasticsearch = require('@quilt/elasticsearch').Elasticsearch;
const infrastructure = require('../../config/infrastructure.js');

const deployment = new quilt.Deployment();
deployment.deploy(infrastructure);
const elasticsearch = new Elasticsearch(infrastructure.nWorker).allowFromPublic();
elasticsearch.deploy(deployment);
