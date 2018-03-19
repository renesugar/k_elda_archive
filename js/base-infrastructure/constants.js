const path = require('path');
const os = require('os');

// This is also defined in bindings.js. This code duplication is ugly, but it
// significantly simplifies packaging the `kelda base-infrastructure` code with the
// "@kelda/install" module.
const infraDirectory = path.join(os.homedir(), '.kelda', 'infra');
const baseInfraLocation = path.join(infraDirectory, 'default.js');

// Keys used in the infrastructure template.
const namespace = 'namespace';
const provider = 'provider';

const size = 'size';
const ram = 'ram';
const cpu = 'cpu';
const preemptible = 'preemptible';
const region = 'region';

const masterCount = 'masterCount';
const workerCount = 'workerCount';

const infraOverwrite = 'infraOverwrite';

// For user defined sizes.
const other = 'other';

module.exports = {
  infraDirectory,
  baseInfraLocation,
  namespace,
  infraOverwrite,
  provider,
  other,
  size,
  ram,
  cpu,
  preemptible,
  region,
  masterCount,
  workerCount,
};
