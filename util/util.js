const path = require('path');
const os = require('os');

const infraDirectory = path.join(os.homedir(), '.quilt', 'infra');

/**
  * Returns the absolute path to the infrastructure with the given name.
  *
  * @param {string} infraName The name of the infrastructure.
  * @return {string} The absolute path to the infrastructure file.
  */
function infraPath(infraName) {
  return path.join(infraDirectory, `${infraName}.js`);
}

let log = console.log;

module.exports = {
  infraDirectory,
  infraPath,
  log,
};
