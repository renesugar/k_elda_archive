const path = require('path');
const fs = require('fs');

const providerFile = path.join(__dirname, 'providers.json');

/**
  * Represents a cloud provider.
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
    const providerInfo = JSON.parse(fs.readFileSync(providerFile, 'utf8'));
    this.sizes = providerInfo[name].sizes || {};
    this.regions = providerInfo[name].regions || {};
    this.hasPreemptible = providerInfo[name].hasPreemptible;
  }
}

module.exports = Provider;
