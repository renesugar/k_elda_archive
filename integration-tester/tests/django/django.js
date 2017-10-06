const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');
const Django = require('@quilt/django');
const haproxy = require('@quilt/haproxy');
const Mongo = require('@quilt/mongo');

const infra = infrastructure.createTestInfrastructure();

const mongo = new Mongo(3);
const django = new Django(3, 'quilt/django-polls', mongo);
const proxy = haproxy.simpleLoadBalancer(django.containers);
proxy.allowFrom(quilt.publicInternet, 80);

django.deploy(infra);
mongo.deploy(infra);
proxy.deploy(infra);
