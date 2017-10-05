const quilt = require('@quilt/quilt');
const infrastructure = require('../../config/infrastructure.js');
const Django = require('@quilt/django');
const haproxy = require('@quilt/haproxy');
const Mongo = require('@quilt/mongo');

const deployment = new quilt.Deployment();
deployment.deploy(infrastructure);

const mongo = new Mongo(3);
const django = new Django(3, 'quilt/django-polls', mongo);
const proxy = haproxy.simpleLoadBalancer(django.containers);
proxy.allowFrom(quilt.publicInternet, 80);

django.deploy(deployment);
mongo.deploy(deployment);
proxy.deploy(deployment);
