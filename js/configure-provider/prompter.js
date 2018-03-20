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
const { Provider, allProviders } = require('./provider');

/**
  * Prompt the user for for their desired provider.
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
  * Ask the user for their provider credentials.
  *
  * @param {Provider} provider The provider to set up credentials for.
  * @returns {Promise} A promise that contains the user's answers.
  */
function credentialsPrompts(provider) {
  if (!provider.requiresCreds()) {
    return new Promise(resolve => resolve({}));
  }

  const keyExists = provider.credsExist();
  const questions = [
    {
      type: 'confirm',
      name: consts.providerUseExistingKey,
      message: `Use existing keys for ${provider.getName()}` +
        ` (${provider.getCredsPath()})?`,
      default: true,
      when() {
        return keyExists;
      },
    },
    {
      type: 'confirm',
      name: consts.credsConfirmOverwrite,
      message: 'This will overwrite the existing credentials. Continue?',
      default: false,
      when(answers) {
        return answers[consts.providerUseExistingKey] === false;
      },
    },
  ];

  const keys = provider.getCredsKeys();
  const keyNames = Object.keys(keys);

  keyNames.forEach((keyName) => {
    questions.push(
      {
        type: 'input',
        name: keyName,
        message: `${keys[keyName]}:`,
        // Ask this question for all credential inputs that are not paths.
        // I.e. keys given exactly as they should appear in the credential file.
        when(answers) {
          return (keyName !== consts.inputCredsPath &&
            (!keyExists || answers[consts.credsConfirmOverwrite]));
        },
        filter(input) {
          return input.trim();
        },
      },
      {
        type: 'input',
        name: keyName,
        message: `${keys[keyName]}:`,

        // Ask this question if the credentials should be given as a file path.
        // E.g. the path to the GCE project ID file.
        when(answers) {
          return keyName === consts.inputCredsPath &&
            (!keyExists || answers[consts.credsConfirmOverwrite]);
        },

        validate(input) {
          if (fs.existsSync(input)) return true;
          return `Oops, no file called "${input}".`;
        },

        filter(input) {
          return path.resolve(input);
        },
      });
  });
  return inquirer.prompt(questions);
}

/**
 * Prompt the user to get the information needed to create their credentials files.
 *
 * @returns {Promise} A promise that contains the user's answers.
 */
function promptUser() {
  const answers = {};
  return providerPrompt()
    .then((providerAnswer) => {
      Object.assign(answers, providerAnswer);
      const provider = new Provider(answers[consts.provider]);
      return credentialsPrompts(provider)

        .then((keyAnswers) => {
          Object.assign(answers, keyAnswers);
          return { provider, answers };
        });
    });
}

module.exports = { promptUser };
