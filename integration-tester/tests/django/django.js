const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');
const Django = require('@kelda/django');
const haproxy = require('@kelda/haproxy');
const Mongo = require('@kelda/mongo');

const infra = infrastructure.createTestInfrastructure();

// Mongo supports at most 7 voting members.
const numMongo = Math.min(7, Math.floor(infrastructure.nWorker / 2));
const mongo = new Mongo(numMongo);
const django = new Django(Math.floor(infrastructure.nWorker / 2), 'keldaio/django-polls', mongo);
const proxy = haproxy.simpleLoadBalancer(django.containers);
proxy.allowFrom(kelda.publicInternet, 80);

django.deploy(infra);
mongo.deploy(infra);
proxy.deploy(infra);
