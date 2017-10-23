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
const name = 'name';
const infraOverwrite = 'infraOverwrite';

// Provider.
const providerUseExistingKey = 'useExistingKey';
const providerCredsPath = 'credsPath';
const credsConfirmOverwrite = 'confirmOverwrite';

// Size.
const other = 'other';

const inputCredsPath = 'credsPath';

module.exports = {
  name,
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
