const {
    Container,
    LabelRule,
    Service,
    createDeployment} = require('@quilt/quilt');
let infrastructure = require('../../config/infrastructure.js');

let deployment = createDeployment({});
deployment.deploy(infrastructure);

let c = new Container('networkstatic/iperf3', ['-s']);

// We want (nWorker - 1) machines with 1 container to test intermachine
// bandwidth. We want 1 machine with 2 containers to test intramachine
// bandwidth. Since inclusive placement is not implemented yet, guarantee
// that one machine has two iperf containers by exclusively placing one
// container on each machine, and then adding one more container to any
// machine.
let exclusive = new Service('iperf', c.replicate(infrastructure.nWorker));
exclusive.place(new LabelRule(true, exclusive));

let extra = new Service('iperfExtra', [c]);

exclusive.allowFrom(exclusive, 5201);
exclusive.allowFrom(extra, 5201);
extra.allowFrom(exclusive, 5201);
exclusive.allowFrom(extra, 5201);

deployment.deploy(exclusive);
deployment.deploy(extra);
