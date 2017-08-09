const quilt = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

let red = new quilt.Service('red', [new quilt.Container('google/pause')]);
deployment.deploy(red);
