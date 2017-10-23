const path = require('path');
const os = require('os');

const keldaSshKeyLocationPrivate = path.join(os.homedir(), '.ssh', 'kelda');
const keldaSshKeyLocationPublic = path.join(os.homedir(), '.ssh', 'kelda.pub');

// Used in infrastructure template.
const namespace = 'namespace';
const provider = 'provider';

const size = 'size';
const ram = 'ram';
const cpu = 'cpu';
const preemptible = 'preemptible';
const region = 'region';

const sshKey = 'sshKey';

const masterCount = 'masterCount';
const workerCount = 'workerCount';


// Infrastructure.
const name = 'name';
const infraOverwrite = 'infraOverwrite';

// Provider.
const providerUseExistingKey = 'useExistingKey';
const providerCredsPath = 'credsPath';
const credsConfirmOverwrite = 'confirmOverwrite';

// SSH.
const sshUseExistingKey = 'Use existing SSH key';
const sshGenerateKey = 'Generate new SSH key pair';
const skip = 'Skip (not recommended)';
const sshKeyOption = 'sshKeyOption';
const sshKeyPath = 'sshKeyPath';

// Size.
const other = 'other';

const inputCredsPath = 'credsPath';

module.exports = {
  keldaSshKeyLocationPrivate,
  keldaSshKeyLocationPublic,

  name,
  namespace,
  infraOverwrite,
  provider,
  providerUseExistingKey,
  providerCredsPath,
  credsConfirmOverwrite,
  sshKey,
  sshUseExistingKey,
  sshGenerateKey,
  skip,
  sshKeyOption,
  sshKeyPath,
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
