const { Infrastructure, Machine } = require('@quilt/quilt');

/** The number of worker machines to launch to run the tests on. */
const nWorker = 3;

/**
 * Creates an Infrastructure to use to run the Quilt integration tests.
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
  return new Infrastructure(baseMachine, workers);
}

module.exports = { createTestInfrastructure, nWorker };
