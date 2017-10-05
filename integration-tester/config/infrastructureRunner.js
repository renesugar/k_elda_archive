const { Deployment } = require('@quilt/quilt');
const infrastructure = require('./infrastructure.js');

(new Deployment()).deploy(infrastructure);
