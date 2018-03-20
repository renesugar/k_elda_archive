const path = require('path');
const exit = require('process').exit;

// Assign imports to variables rather than consts so the modules can
// be mocked out in tests.
/* eslint-disable prefer-const */
let fs = require('fs');
let fsExtra = require('fs-extra');

let log = console.log;
/* eslint-enable prefer-const */

const handlebars = require('handlebars');

const prompter = require('./prompter');
const consts = require('./constants');

const templateDir = path.join(__dirname, 'templates');

/**
* Create a new provider credentials file if needed.
*
* @param {Provider} provider The chosen provider.
* @param {Promise} answers A promise that contains the user's answers.
* @returns {void}
  */
function processAnswers(provider, answers) {
  if (!provider.requiresCreds()) {
    log(`${provider.getName()} doesn't require any credentials. You're good to go!`);
    return;
  }

  if (answers[consts.providerUseExistingKey] ||
      answers[consts.credsConfirmOverwrite] === false) {
    log(`Ok! Kelda will use your existing credentials for ${provider.getName()}.`);
    return;
  }

  const credentialsDest = provider.getCredsPath();
  try {
    fsExtra.mkdirpSync(path.dirname(credentialsDest));
  } catch (err) {
    throw new Error(
      `failed to create credentials file in ${credentialsDest}: ${err}`);
  }

  if (provider.credsTemplate !== undefined) {
    const templateFile = path.join(templateDir, provider.getCredsTemplate());
    if (!fs.existsSync(templateFile)) {
      throw new Error(`failed to create file at ${credentialsDest}.` +
        ` No template in ${templateFile}.`);
    }

    const rawTemplate = fs.readFileSync(templateFile, 'utf8');
    const template = handlebars.compile(rawTemplate);
    fs.writeFileSync(credentialsDest, template(answers));
  } else {
    // We were given a file path instead of the keys themselves.
    const credentialsSrc = answers[consts.inputCredsPath];
    fsExtra.copySync(credentialsSrc, credentialsDest);
  }

  log(`Wrote ${provider.getName()} credentials to ${credentialsDest}.`);
}

/**
  * Get user input, and process the answers.
  * @returns {void}
  */
function run() {
  prompter.promptUser()
    .then((result) => {
      if (result.shouldAbort) { exit(0); }
      processAnswers(result.provider, result.answers);
    })
    .catch((error) => {
      console.warn(`Oops, something went wrong: ${error.message}`);
      exit(1);
    });
}

if (require.main === module) {
  log('-------------------------------------------------------------\n' +
    '|    Kelda needs access to your cloud provider credentials    |\n' +
    '|    in order to launch VMs in your account. See details at   |\n' +
    '|    http://docs.kelda.io/#cloud-provider-configuration.      |\n' +
    '--------------------------------------------------------------');
  run();
}

module.exports = { processAnswers };
