const kelda = require('kelda');
const haproxy = require('@kelda/haproxy');
const Mongo = require('@kelda/mongo');
const Node = require('@kelda/nodejs');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

// Mongo supports at most 7 voting members.
const numMongo = Math.min(7, Math.floor(infrastructure.nWorker / 2));
const mongo = new Mongo(numMongo);
const app = new Node({
  nWorker: Math.floor(infrastructure.nWorker / 2),
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
