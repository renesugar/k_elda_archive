// Disable requiring FunctionExpression JSDoc, because this file
// uses many function expressions as lambda expressions where
// it is overkill to require documentation.
/*
eslint "require-jsdoc": ["error", {
    "require": {
        "FunctionExpression": false
    }
}]
*/
const path = require('path');
const fs = require('fs');

const inquirer = require('inquirer');
const consts = require('./constants');
const Provider = require('./provider');

/**
  * Return a list of all provider names as listed in providers.json.
  *
  * @returns {string[]} A list of provider names.
  */
function allProviders() {
  const providerFile = path.join(__dirname, 'providers.json');
  const providerInfo = JSON.parse(fs.readFileSync(providerFile, 'utf8'));
  return Object.keys(providerInfo);
}

/**
  * Throw an error if the input string does not contain a number.
  *
  * @param {string} input The string to check.
  * @returns {void}
  */
function isNumber(input) {
  if (!(/^\d+$/.test(input))) {
    throw new Error('Please provide a number');
  }
  return true;
}

/**
  * Converts a map from friendly names to formal values into an array of objects
  * compatible with the `choices` attribute of inquirer questions. For example,
  * `{small: 't2.micro'}` would convert to
  * `[{name: 'small (t2.micro)', value: 't2.micro'}].
  *
  * @param {Object.<string, string>} friendlyNameToValue - The data to make descriptions for.
  * @returns {Object.<string, string>[]} An array of {name, val} objects.
  *
  */
function getInquirerDescriptions(friendlyNameToValue) {
  return Object.keys(friendlyNameToValue).map((friendlyName) => {
    const formalValue = friendlyNameToValue[friendlyName];
    return { name: `${friendlyName} (${formalValue})`, value: formalValue };
  });
}

// User prompts.

/**
  * If a base infrastructure exists, ask the user if they want to overwrite it.
  *
  * @returns {Promise} A promise that contains the user's answers.
  */
function overwritePrompt() {
  const questions = [{
    type: 'confirm',
    name: consts.infraOverwrite,
    message: 'This will overwrite your existing base infrastructure. Continue?',
    default: false,
    when() {
      return fs.existsSync(consts.baseInfraLocation);
    },
  },
  ];
  return inquirer.prompt(questions);
}

/**
  * Prompt the user for their desired provider.
  *
  * @returns {Promise} A promise that contains the user's answer.
  */
function providerPrompt() {
  const questions = [
    {
      type: 'list',
      name: consts.provider,
      message: 'Choose a provider:',
      choices() {
        return allProviders();
      },
    },
  ];
  return inquirer.prompt(questions);
}

/**
  * Ask the user for machine configuration, such as size and region.
  *
  * @param {Provider} provider The provider chosen for this infrastructure.
  * @returns {Promise} A promise that contains the user's answers.
  */
function machineConfigPrompts(provider) {
  const regionChoices = getInquirerDescriptions(provider.regions);
  const sizeChoices = getInquirerDescriptions(provider.sizes);
  sizeChoices.push({ name: consts.other, value: consts.other });

  const questions = [
    {
      type: 'confirm',
      name: consts.preemptible,
      message: 'Do you want to run preemptible instances?',
      default: false,
      when() {
        return provider.hasPreemptible;
      },
    },
    {
      type: 'list',
      name: consts.region,
      message: 'Which region do you want to deploy in?',
      choices: regionChoices,
      when() {
        return Object.keys(provider.regions).length !== 0;
      },
    },
    {
      type: 'list',
      name: consts.size,
      message: 'What machine size do you want?',
      choices: sizeChoices,
      when() {
        return Object.keys(provider.sizes).length !== 0;
      },
    },
    {
      type: 'input',
      name: consts.size,
      message: 'Which other instance type?',
      validate(input) {
        if (input !== '') return true;
        return 'Please provide an instance type';
      },
      when(answers) {
        return answers[consts.size] === consts.other;
      },
    },
    {
      type: 'input',
      name: consts.cpu,
      message: 'How many CPUs do you want?',
      default: 1,
      validate(input) {
        return isNumber(input);
      },
      when() {
        return Object.keys(provider.sizes).length === 0;
      },
    },
    {
      type: 'input',
      name: 'ram',
      message: 'How many GiB of RAM do you want?',
      default: 2,
      validate(input) {
        return isNumber(input);
      },
      when() {
        return Object.keys(provider.sizes).length === 0;
      },
    },
  ];
  return inquirer.prompt(questions);
}

/**
  * Ask for the desired number of master machines.
  *
  * @returns {Promise} A promise that contains the user's answer.
  */
function masterMachinePrompts() {
  const questions = [
    {
      type: 'input',
      name: consts.masterCount,
      message: 'How many master VMs?',
      default: 1,
      validate(input) {
        return isNumber(input);
      },
    },
  ];
  console.log('Master machines are in charge of keeping your application running. ' +
  'Most users just need 1 master, but more can be added for fault tolerance.');
  return inquirer.prompt(questions);
}

/**
  * Ask for the desired number of worker machines.
  *
  * @returns {Promise} A promise that contains the user's answer.
  */
function workerMachinePrompts() {
  const questions = [
    {
      type: 'input',
      name: consts.workerCount,
      message: 'How many worker VMs?',
      default: 1,
      validate(input) {
        return isNumber(input);
      },
    },
  ];
  console.log('Worker VMs run your application code. For small applications, 1 ' +
  'worker is typically enough. For applications with many containers, you probably ' +
  'want more.');
  return inquirer.prompt(questions);
}

/**
 * Prompt the user for the information needed to create a new base
 * infrastructure.
 *
 * @returns {Promise} A promise that contains the user's answers.
 */
function promptUser() {
  const answers = {};
  return overwritePrompt()
    .then((shouldOverwrite) => {
      if (shouldOverwrite[consts.infraOverwrite] === false) {
        return { shouldAbort: true };
      }
      return providerPrompt()
        .then((providerAnswer) => {
          Object.assign(answers, providerAnswer);
          const provider = new Provider(answers[consts.provider]);
          return machineConfigPrompts(provider)

            .then((machineConfigAnswers) => {
              Object.assign(answers, machineConfigAnswers);
              return masterMachinePrompts();
            })

            .then((masterAnswers) => {
              Object.assign(answers, masterAnswers);
              return workerMachinePrompts();
            })

            .then((workerAnswers) => {
              Object.assign(answers, workerAnswers);
              return { answers };
            });
        });
    });
}

module.exports = {
  promptUser,
  allProviders,
  isNumber,
  getInquirerDescriptions,
};
