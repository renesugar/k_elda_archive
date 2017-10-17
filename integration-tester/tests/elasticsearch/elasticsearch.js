const Elasticsearch = require('@kelda/elasticsearch').Elasticsearch;
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();
const elasticsearch = new Elasticsearch(infrastructure.nWorker).allowFromPublic();
elasticsearch.deploy(infra);
