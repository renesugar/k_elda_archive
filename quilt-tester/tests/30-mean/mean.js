const quilt = require('@quilt/quilt');
let haproxy = require('@quilt/haproxy');
let Mongo = require('@quilt/mongo');
let Node = require('@quilt/nodejs');
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

let mongo = new Mongo(3);
let app = new Node({
  nWorker: 3,
  repo: 'https://github.com/tejasmanohar/node-todo.git',
  env: {
    PORT: '80',
    MONGO_URI: mongo.uri('mean-example'),
  },
});

let proxy = haproxy.simpleLoadBalancer(app.cluster);

mongo.allowFrom(app.cluster, mongo.port);
proxy.allowFrom(quilt.publicInternet, haproxy.exposedPort);

deployment.deploy([app, mongo, proxy]);
