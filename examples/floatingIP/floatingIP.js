// AWS: http://docs.aws.amazon.com/cli/latest/reference/ec2/allocate-address.html
// Google: https://cloud.google.com/compute/docs/configure-instance-ip-addresses#reserve_new_static
const { Deployment, Machine } = require('@quilt/quilt');
const nginx = require('@quilt/nginx');

const floatingIp = 'xxx.xxx.xxx.xxx (CHANGE ME)';
const deployment = new Deployment();

const app = nginx.createContainer(80);

app.placeOn({ floatingIp });
app.deploy(deployment);

const baseMachine = new Machine({
  provider: 'Amazon',
  size: 'm4.large',
  region: 'us-west-2',
  // sshKeys: githubKeys("GITHUB_USERNAME")
});

deployment.deploy(baseMachine.asMaster());

baseMachine.floatingIp = floatingIp;
deployment.deploy(baseMachine.asWorker());
