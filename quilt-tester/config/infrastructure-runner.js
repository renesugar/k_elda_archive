const {createDeployment} = require('@quilt/quilt');
let infrastructure = require('./infrastructure.js');

createDeployment({}).deploy(infrastructure);
