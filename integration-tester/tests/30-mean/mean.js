const kelda = require('kelda');
const haproxy = require('@kelda/haproxy');
const Mongo = require('@kelda/mongo');
const Node = require('@kelda/nodejs');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const mongo = new Mongo(3);
const app = new Node({
  nWorker: 3,
  repo: 'https://github.com/kelda/node-todo.git',
  env: {
    PORT: '80',
    MONGO_URI: mongo.uri('mean-example'),
  },
});

const proxy = haproxy.simpleLoadBalancer(app.containers);

mongo.allowFrom(app.containers, mongo.port);
proxy.allowFrom(kelda.publicInternet, haproxy.exposedPort);

app.deploy(infra);
mongo.deploy(infra);
proxy.deploy(infra);
