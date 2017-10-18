const path = require('path');
const os = require('os');

// Both infraDirectory and infraPath are also defined in bindings.js.
// This code duplication is ugly, but it significantly simplifies packaging
// the `kelda init` code with the "@kelda/install" module.
const infraDirectory = path.join(os.homedir(), '.kelda', 'infra');

/**
  * Returns the absolute path to the infrastructure with the given name.
  *
  * @param {string} infraName The name of the infrastructure.
  * @returns {string} The absolute path to the infrastructure file.
  */
function infraPath(infraName) {
  return path.join(infraDirectory, `${infraName}.js`);
}

module.exports = {
  infraDirectory,
  infraPath,
};
