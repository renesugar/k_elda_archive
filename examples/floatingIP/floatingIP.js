// AWS: http://docs.aws.amazon.com/cli/latest/reference/ec2/allocate-address.html
// Google: https://cloud.google.com/compute/docs/configure-instance-ip-addresses#reserve_new_static
const { Infrastructure, Machine } = require('kelda');
const nginx = require('@kelda/nginx');

const floatingIp = 'xxx.xxx.xxx.xxx (CHANGE ME)';

const baseMachine = new Machine({
  provider: 'Amazon',
  size: 'm4.large',
  region: 'us-west-2',
  // sshKeys: githubKeys("GITHUB_USERNAME")
});
const workerMachine = baseMachine.clone();
workerMachine.floatingIp = floatingIp;

const infra = new Infrastructure({ masters: baseMachine, workers: workerMachine });

const app = nginx.createContainer(80);

app.placeOn({ floatingIp });
app.deploy(infra);
