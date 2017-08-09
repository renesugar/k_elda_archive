const quilt = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = quilt.createDeployment();
deployment.deploy(infrastructure);

let c = new quilt.Container('networkstatic/iperf3', ['-s']);

// If we deploy nWorker+1 containers, at least one machine is guaranteed to run
// two containers, and thus be able to test intra-machine bandwidth.
let iperfs = new quilt.Service('iperf', c.replicate(infrastructure.nWorker+1));
iperfs.allowFrom(iperfs, 5201);
deployment.deploy(iperfs);
