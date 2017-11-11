const path = require('path');
const os = require('os');

// This is also defined in bindings.js. This code duplication is ugly, but it
// significantly simplifies packaging the `kelda init` code with the
// "@kelda/install" module.
const infraDirectory = path.join(os.homedir(), '.kelda', 'infra');
const baseInfraLocation = path.join(infraDirectory, 'default.js');

// Used in infrastructure template.
const namespace = 'namespace';
const provider = 'provider';

const size = 'size';
const ram = 'ram';
const cpu = 'cpu';
const preemptible = 'preemptible';
const region = 'region';

const masterCount = 'masterCount';
const workerCount = 'workerCount';


// Infrastructure.
const infraOverwrite = 'infraOverwrite';

// Provider.
const providerUseExistingKey = 'useExistingKey';
const providerCredsPath = 'credsPath';
const credsConfirmOverwrite = 'confirmOverwrite';

// Size.
const other = 'other';

const inputCredsPath = 'credsPath';

module.exports = {
  infraDirectory,
  baseInfraLocation,
  namespace,
  infraOverwrite,
  provider,
  providerUseExistingKey,
  providerCredsPath,
  credsConfirmOverwrite,
  other,
  size,
  ram,
  cpu,
  preemptible,
  region,
  masterCount,
  workerCount,
  inputCredsPath,
};
