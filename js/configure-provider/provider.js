const path = require('path');
const fs = require('fs');
const os = require('os');

const consts = require('./constants');

// The credentials key names (e.g. 'key' and 'secret') should correspond to the
// keys used in the provider's credentials template.
// If the credentials should be given as a file path, the key name should be
// [consts.inputCredsPath].
let credentialsInfo = { // eslint-disable-line prefer-const
  Amazon: {
    credsTemplate: 'amazon_creds_template',
    credsKeys: { key: 'AWS access key id', secret: 'AWS secret access key' },
    credsLocation: ['.aws', 'credentials'],
  },
  Google: {
    credsKeys: { [consts.inputCredsPath]: 'Path to GCE service account key' },
    credsLocation: ['.gce', 'kelda.json'],
  },
  DigitalOcean: {
    credsTemplate: 'digitalocean_creds_template',
    credsKeys: { key: 'DigitalOcean account token' },
    credsLocation: ['.digitalocean', 'key'],
  },
  Vagrant: {},
};

/**
  * Return a list of all available provider names.
  *
  * @returns {string[]} A list of provider names.
  */
function allProviders() {
  return Object.keys(credentialsInfo);
}

/**
  * Represents a cloud provider.
  */
class Provider {
  /**
    * Creates a new Provider instance.
    *
    * @param {string} name The desired name of the provider. Should match the
    *   name in credentialsInfo.
    */
  constructor(name) {
    this.name = name;
    this.credsTemplate = credentialsInfo[name].credsTemplate;
    this.credsKeys = credentialsInfo[name].credsKeys;
    this.credsLocation = credentialsInfo[name].credsLocation;
  }

  /**
    * @returns {string} The name of this provider.
    */
  getName() {
    return this.name;
  }

  /**
    * @returns {Object.<string, string>} A map where the keys are the keys needed
    *   by this provider's credentials template, and the values are the user friendly
    *   descriptions.
    */
  getCredsKeys() {
    return this.credsKeys || {};
  }

  /**
    * Returns the name of the file that contains the credentials template.
    * Note, this is only the file name, not the entire path.
    *
    * @returns {string} The file name.
    */
  getCredsTemplate() {
    return this.credsTemplate || '';
  }

  /**
    * @returns {boolean} True if this provider requires credentials, else false.
    */
  requiresCreds() {
    return this.credsLocation !== undefined;
  }

  /**
    * @returns {boolean} True if there exist credentials for this provider in
    *   the default location, otherwise false.
    */
  credsExist() {
    return fs.existsSync(this.getCredsPath());
  }

  /**
    * If the provider uses credentials, return the path where the credentials
    * should be.
    *
    * @returns {string} An absolute path to the credentials, if needed. Otherwise
    *   an empty string.
    */
  getCredsPath() {
    if (this.requiresCreds()) {
      const pathArray = [os.homedir()].concat(this.credsLocation);
      return path.join(...pathArray);
    }
    return '';
  }
}

module.exports = { Provider, allProviders };
