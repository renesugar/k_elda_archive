#!/usr/bin/env node

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
const infraTemplateFile = path.join(templateDir, 'inf_template');

/**
  * Creates a file from the given template and context, and writes the result to
  * a file.
  *
  * @param {string} templateLocation The path to the template file.
  * @param {Object.<string, string>} context The context to apply to the
  *   template.
  * @param {string} destination The path to write the result to.
  * @returns {void}
  */
function createFileFromTemplate(templateLocation, context, destination) {
  if (!fs.existsSync(templateLocation)) {
    throw new Error(`failed to create file at ${destination}.` +
      ` No template in ${templateLocation}.`);
  }
  const rawTemplate = fs.readFileSync(templateLocation, 'utf8');
  const template = handlebars.compile(rawTemplate);
  fs.writeFileSync(destination, template(context));
}

/**
  * Create new credentials file if needed.
  *
  * @param {Provider} provider The chosen provider.
  * @param {Promise} answers A promise that contains the user's answers.
  * @returns {void}
  */
function writeProviderCreds(provider, answers) {
  if (!provider.requiresCreds() ||
      answers[consts.providerUseExistingKey] ||
      answers[consts.credsConfirmOverwrite] === false) {
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
    createFileFromTemplate(templateFile, answers, credentialsDest);
  } else {
    // We were given a file path instead of the keys themselves.
    const credentialsSrc = answers[consts.inputCredsPath];
    fsExtra.copySync(credentialsSrc, credentialsDest);
  }

  log(`Wrote credentials to ${credentialsDest}`);
}

/** Create the infrastructure file and credentials based on the user's input.
  *
  * @param {Provider} provider The chosen provider.
  * @param {Promise} answers A promise that contains the user's answers.
  * @returns {void}
  */
function processAnswers(provider, answers) {
  try {
    fsExtra.mkdirpSync(consts.infraDirectory);
  } catch (err) {
    throw new Error(`failed to create ${consts.infraDirectory}: ${err}`);
  }

  writeProviderCreds(provider, answers);

  const config = {
    [consts.provider]: answers[consts.provider],
    [consts.size]: answers[consts.size],
    [consts.ram]: answers[consts.ram],
    [consts.cpu]: answers[consts.cpu],
    [consts.preemptible]: answers[consts.preemptible] || false,
    [consts.region]: answers[consts.region],
    [consts.masterCount]: answers[consts.masterCount],
    [consts.workerCount]: answers[consts.workerCount],
  };

  createFileFromTemplate(infraTemplateFile, config, consts.baseInfraLocation);
  log('The base infrastructure has been created! Use ' +
    '`baseInfrastructure() in your blueprint to use this infrastructure.');
}

/**
  * Get user input, and process the answers.
  * @returns {void}
  */
function run() {
  prompter.promptUser()
    .then((result) => {
      if (result.shouldAbort) {
        exit(0);
      }
      processAnswers(result.provider, result.answers);
    })
    .catch((error) => {
      console.warn(`Oops, something went wrong: ${error.message}`);
      exit(1);
    });
}

if (require.main === module) {
  console.log(`---------------------------------------------------------
|   See docs at http://docs.kelda.io/#init              |
---------------------------------------------------------`);
  run();
}

module.exports = {
  createFileFromTemplate,
  writeProviderCreds,
  processAnswers,
};
