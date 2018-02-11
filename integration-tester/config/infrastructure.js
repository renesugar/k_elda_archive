const { Infrastructure, Machine } = require('kelda');

/** The number of worker machines to launch to run the tests on. */
const nWorker = getNumberWorkers();

const providerToFloatingIp = {
  Amazon: '13.57.99.49', // us-west-1
  Google: '104.196.11.66', // us-east1-b
  DigitalOcean: '165.227.242.236', // sfo2
};

/**
 * getNumberWorkers returns the number of Kelda workers to boot by parsing
 * the NUMBER_WORKERS environment variable, and defaulting to 3 if it is not
 * defined.
 *
 * @returns {int} The number of workers to boot.
 */
function getNumberWorkers() {
  if (process.env.NUMBER_WORKERS !== undefined) {
    return parseInt(process.env.NUMBER_WORKERS, 10);
  }
  return 3;
}

/**
 * Creates an Infrastructure to use to run the Kelda integration tests.
 *
 * @returns {Infrastructure} An Infrastructure to use to run the tests.
 */
function createTestInfrastructure() {
  const baseMachine = new Machine({
    provider: process.env.PROVIDER || 'Amazon',
    size: process.env.SIZE || 'm3.medium',
    sshKeys: ['ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCxMuzNUdKJREFgUkS' +
      'pD0OPjtgDtbDvHQLDxgqnTrZpSvTw5r8XDd+AFS6eVibBfYv1u+geNF3IEkpO' +
      'klDlII37DzhW7wzlRB0SmjUtODxL5hf9hKoScDpvXG3RBD6PBCyOHA5IJBTqP' +
      'GpIZUMmOlXDYZA1KLaKQs6GByg7QMp6z1/gLCgcQygTDdiTfESgVMwR1uSQ5M' +
      'RjBaL7vcVfrKExyCLxito77lpWFMARGG9W1wTWnmcPrzYR7cLzhzUClakazNJ' +
      'mfso/b4Y5m+pNH2dLZdJ/eieLtSEsBDSP8X0GYpmTyFabZycSXZFYP+wBkrUT' +
      'mgIh9LQ56U1lvA4UlxHJ'],
  });

  const workers = [];
  for (let i = 0; i < nWorker; i += 1) {
    const worker = baseMachine.clone();
    // Preemptible instances are currently only implemented on Amazon.
    if (baseMachine.provider === 'Amazon') {
      worker.preemptible = true;
    }
    workers.push(worker);
  }
  workers[0].floatingIp = providerToFloatingIp[baseMachine.provider];

  return new Infrastructure({
    masters: baseMachine,
    workers,
    namespace: process.env.TESTING_NAMESPACE,
  });
}

module.exports = { createTestInfrastructure, nWorker };
