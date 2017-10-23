const path = require('path');
const fs = require('fs');
const os = require('os');

const consts = require('./constants');

const providerFile = path.join(__dirname, 'providers.json');

// The credentials key names (e.g. 'key' and 'secret') should correspond to the
// keys used in the provider's credentials template.
// If the credentials should be given as a file path, the key name should be
// [inputCredsPath].
let providerConfig = { // eslint-disable-line prefer-const
  Amazon: {
    credsTemplate: 'amazon_creds_template',
    credsKeys: {
      key: 'AWS access key id',
      secret: 'AWS secret access key',
    },
  },
  Google: {
    credsKeys: {
      [consts.inputCredsPath]: 'Path to GCE service account key',
    },
  },
  DigitalOcean: {
    credsTemplate: 'digitalocean_creds_template',
    credsKeys: {
      key: 'DigitalOcean account token',
    },
  },
  Vagrant: {},
};

/**
  * Represents a Provider.
  */
class Provider {
  /**
    * Constructs a new Provider instance.
    *
    * @param {string} name The desired name of the provider. Should match the
    *   name in providers.json.
    */
  constructor(name) {
    this.name = name;
    this.credsTemplate = providerConfig[name].credsTemplate;
    this.credsKeys = providerConfig[name].credsKeys;

    const providerInfo = JSON.parse(fs.readFileSync(providerFile, 'utf8'));

    this.sizes = providerInfo[name].sizes;
    this.regions = providerInfo[name].regions;
    this.hasPreemptible = providerInfo[name].hasPreemptible;
    this.credsLocation = providerInfo[name].credsLocation;
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
    * @returns {Object.<string, string>} An object with suggested sizes for this
    *   provider. The keys are user friendly descriptions (e.g. 'small') of the
    *   size, and the values are the actual size names used by the provider.
    */
  getSizes() {
    return this.sizes || {};
  }

  /**
    * @returns {Object.<string, string>} An object with supported sizes for this
    * provider. The keys are user friendly descriptions (e.g. 'N. California')
    * of the region, and the values are the actual region names used by the
    * provider (us-west-1).
    */
  getRegions() {
    return this.regions || {};
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

module.exports = Provider;
