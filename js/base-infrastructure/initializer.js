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

const infraTemplateFile = path.join(__dirname, 'infrastructure.js.tmpl');

/** Create the base infrastructure file based on the user's input.
  *
  * @param {Promise} answers A promise that contains the user's answers.
  * @returns {void}
  */
function processAnswers(answers) {
  try {
    fsExtra.mkdirpSync(consts.infraDirectory);
  } catch (err) {
    throw new Error(`failed to create ${consts.infraDirectory}: ${err}`);
  }

  if (!fs.existsSync(infraTemplateFile)) {
    throw new Error('failed to create base infrastructure.' +
      ` No template in ${infraTemplateFile}.`);
  }
  const rawTemplate = fs.readFileSync(infraTemplateFile, 'utf8');
  const template = handlebars.compile(rawTemplate);
  fs.writeFileSync(consts.baseInfraLocation, template(answers));

  log('The base infrastructure has been created! Call baseInfrastructure() in ' +
    'your blueprint to use this infrastructure.');
}

/**
  * Get user input, and process the answers.
  * @returns {void}
  */
function run() {
  prompter.promptUser()
    .then((result) => {
      if (result.shouldAbort) { exit(0); }
      processAnswers(result.answers);
    })
    .catch((error) => {
      console.warn(`Oops, something went wrong: ${error.message}`);
      exit(1);
    });
}

if (require.main === module) {
  run();
}

module.exports = { processAnswers };
