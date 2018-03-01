const kelda = require('kelda');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

const volume = new kelda.Volume({
  name: 'docker',
  type: 'hostPath',
  path: '/var/run/docker.sock',
});

const noMountArgs = {
  name: 'no-mount',
  image: 'ubuntu',
  command: ['tail', '-f', '/dev/null'],
};

const hasMountArgs = Object.assign({}, noMountArgs, {
  name: 'has-mount',
  volumeMounts: [
    new kelda.VolumeMount({
      volume,
      mountPath: '/docker.sock',
    }),
  ],
});

for (let i = 0; i < infra.workers.length; i += 1) {
  (new kelda.Container(noMountArgs)).deploy(infra);
  (new kelda.Container(hasMountArgs)).deploy(infra);
}
