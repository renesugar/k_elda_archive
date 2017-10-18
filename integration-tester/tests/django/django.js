const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');
const Django = require('@kelda/django');
const haproxy = require('@kelda/haproxy');
const Mongo = require('@kelda/mongo');

const infra = infrastructure.createTestInfrastructure();

const mongo = new Mongo(3);
const django = new Django(3, 'keldaio/django-polls', mongo);
const proxy = haproxy.simpleLoadBalancer(django.containers);
proxy.allowFrom(kelda.publicInternet, 80);

django.deploy(infra);
mongo.deploy(infra);
proxy.deploy(infra);
