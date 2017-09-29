const path = require('path');
const os = require('os');
const fs = require('fs');

const inquirer = require('inquirer');
const consts = require('./constants');
const util = require('./init-util');
const Provider = require('./provider');


/**
  * Return a list of all provider names as listed in providers.json.
  *
  * @return {string[]} A list of provider names.
  */
function allProviders() {
  const providerFile = path.join(__dirname, 'providers.json');
  const providerInfo = JSON.parse(fs.readFileSync(providerFile, 'utf8'));
  return Object.keys(providerInfo);
}
/**
  * Check if the Quilt SSH keys exists.
  *
  * @return {boolean} True iff both the private and public Quilt SSH keys exist.
  */
function quiltSshKeyExists() {
  return (fs.existsSync(consts.quiltSshKeyLocationPublic) &&
      fs.existsSync(consts.quiltSshKeyLocationPrivate));
}

/**
  * Throw an error if the input string does not contain a number.
  *
  * @param {string} input The string to check.
  * @return {void}
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
  * @param {Object.<string, string>} info The data to make descriptions for.
  * @return {Object.<string, string>[]} An array of {name, val} objects.
  *
  */
function getInquirerDescriptions(friendlyNameToValue) {
  return Object.keys(friendlyNameToValue).map((friendlyName) => {
    const formalValue = friendlyNameToValue[friendlyName];
    return { name: `${friendlyName} (${formalValue})`, value: formalValue };
  });
}

function questionWithHelp(question, helpstring) {
  const newQuestion = Object.assign({}, question);

  newQuestion.message = `(type ? for help) ${question.message}`;
  newQuestion.validate = function validate(input) {
    if (input === '?') {
      return helpstring;
    }
    if (question.validate === undefined) {
      return true;
    }
    return question.validate(input);
  };
  return newQuestion;
}

// User prompts.

/**
  * Prompt the user for a valid infrastructure name. Users will use this name in
  * their blueprints to get the right infrastructure.
  *
  * @return {Promise} A promise that contains the user's answers.
  */
function infraNamePrompt() {
  const questions = [
    questionWithHelp({
      type: 'input',
      name: consts.name,
      message: 'Pick a name for your infrastructure:',
      default: 'default',
    }, 'The infrastructure name is used in blueprints to retrieve the ' +
      'infrastructure. E.g. if you choose the name \'aws-small\', then ' +
      '`baseInfrastructure(\'aws-small\')` will return this infrastructure.'),
    {
      type: 'confirm',
      name: consts.infraOverwrite,
      message(answers) {
        return `"${answers[consts.name]}" already exists. Overwrite?`;
      },
      default: false,
      when(answers) {
        return fs.existsSync(util.infraPath(answers[consts.name]));
      },
    },
  ];

  /**
    * Repeatedly prompt for a name until we get a valid answer.
    *
    * @return {Promise} A promise that contains the user's answers
    */
  function confirmIfOverwrite() {
    return inquirer.prompt(questions).then((answers) => {
      if (answers[consts.infraOverwrite] === false) {
        return confirmIfOverwrite();
      }
      return answers;
    });
  }
  return confirmIfOverwrite();
}

/**
  * Prompt the user for for their desired provider.
  *
  * @return {Promise} A promise that contains the user's answer.
  */
function getProviderPrompt() {
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
  * Ask the user about SSH keys for the infrastructure.
  *
  * @param {Provider} provider The chosen provider.
  * @return {Promise} A promise that contains the user's answers.
  */
function sshKeyPrompts(provider) {
  if (!provider.requiresSsh) return { [consts.sshKeyOption]: consts.skip };

  let choices = [consts.sshUseExistingKey, consts.skip];
  if (!quiltSshKeyExists()) {
    choices = [consts.sshGenerateKey].concat(choices);
  }

  const questions = [
    {
      type: 'list',
      name: consts.sshKeyOption,
      message: 'Choose an SSH key to log into VMs and containers:',
      choices,
    },
    {
      type: 'input',
      name: consts.sshKeyPath,
      message: 'Path to public SSH key:',

      when(answers) {
        return answers[consts.sshKeyOption] === consts.sshUseExistingKey;
      },

      default() {
        if (quiltSshKeyExists()) return consts.quiltSshKeyLocationPublic;
        return path.join(os.homedir(), '.ssh', 'id_rsa.pub');
      },

      validate(input) {
        if (fs.existsSync(input)) return true;
        return `Oops, no file called ${input}`;
      },
    },
  ];

  return inquirer.prompt(questions);
}

/**
  * Ask the user about provider credentials.
  *
  * @param {Provider} provider The provider to set up credentials for.
  * @return {Promise} A promise that contains the user's answers.
  */
function credentialsPrompts(provider) {
  if (!provider.requiresCreds()) {
    return new Promise((resolve) => {
      resolve({});
    });
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
      message: 'This will overwrite the existing credentials. Continue?' +
        ' (If no, use existing credentials)',
      default: false,
      when(answers) {
        return answers[consts.providerUseExistingKey] === false;
      },
    },
  ];

  const keys = provider.getCredsKeys();
  const keyNames = Object.keys(keys);

  const credsHelp = `Quilt needs access to your ${provider.getName()} ` +
  'credientials in order to launch VMs in your account. See details at ' +
  'http://docs.quilt.io/#cloud-provider-configuration';

  keyNames.forEach((keyName) => {
    questions.push(
      questionWithHelp({
        type: 'input',
        name: keyName,
        message: `${keys[keyName]}:`,
        // Ask this questions for all credential inputs that are not paths.
        // I.e. keys given exactly as they should appear in the credential file.
        when(answers) {
          return (keyName !== consts.inputCredsPath &&
            (!keyExists || answers[consts.credsConfirmOverwrite]));
        },
        filter(input) {
          return input.trim();
        },
      }, credsHelp),
      questionWithHelp({
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
      }, credsHelp));
  });
  return inquirer.prompt(questions);
}

/**
  * Ask the user for machine configuration, such as size and region.
  *
  * @param {Provider} provider The provider chosen for this infrastructure.
  * @return {Promise} A promise that contains the user's answers.
  */
function machineConfigPrompts(provider) {
  const regionChoices = getInquirerDescriptions(provider.getRegions());
  const sizeChoices = getInquirerDescriptions(provider.getSizes());
  sizeChoices.push({ name: consts.other, value: consts.other });

  const sizeOptions = [
    {
      name: `${consts.instanceTypeSize} (recommended)`,
      value: consts.instanceTypeSize,
    },
    { name: consts.ramCpuSize },
  ];

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
        return provider.regions !== undefined;
      },
    },
    {
      type: 'list',
      name: consts.sizeType,
      message: 'How do you want to specify the size of your machine?',
      choices: sizeOptions,
      when() {
        // If there are no instance types, just ask for RAM and CPU.
        return provider.sizes !== undefined;
      },
    },
    {
      type: 'list',
      name: consts.size,
      message: 'Choose an instance:',
      choices: sizeChoices,
      when(answers) {
        return answers[consts.sizeType] === consts.instanceTypeSize;
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
    questionWithHelp({
      type: 'input',
      name: consts.cpu,
      message: 'How many CPUs do you want?',
      default: 1,
      validate(input) {
        return isNumber(input);
      },
      when(answers) {
        return answers[consts.sizeType] !== consts.instanceTypeSize;
      },
    }, 'For small applications, 1 CPU is probably enough.'),
    questionWithHelp({
      type: 'input',
      name: 'ram',
      message: 'How many GiB of RAM do you want?',
      default: 2,
      validate(input) {
        return isNumber(input);
      },
      when(answers) {
        return answers[consts.sizeType] !== consts.instanceTypeSize;
      },
    }, 'For small applications, 2 GiB is a suitable choice.'),
  ];

  return inquirer.prompt(questions);
}

/**
  * Ask for the desired number of machines.
  *
  * @return {Promise} A promise that contains the user's answers.
  */
function machineCountPrompts() {
  const questions = [
    questionWithHelp({
      type: 'input',
      name: consts.masterCount,
      message: 'How many master VMs?',
      default: 1,
      validate(input) {
        return isNumber(input);
      },
    }, 'Master VMs are in charge of keeping your application running. Most ' +
    'users just need 1, but more can be added for fault tolerance.'),
    questionWithHelp({
      type: 'input',
      name: consts.workerCount,
      message: 'How many worker VMs?',
      default: 1,
      validate(input) {
        return isNumber(input);
      },
    }, 'Worker VMs run your application code. For small applications, 1 ' +
    'worker is typically enough.'),
  ];

  return inquirer.prompt(questions);
}

/**
 * Prompt the user to get the information needed to create the new
 * infrastructure.
 *
 * @return {Promise} A promise that contains the user's answers.
 */
function promptUser() {
  const answers = {};
  return infraNamePrompt()
    .then((infraAnswers) => {
      Object.assign(answers, infraAnswers);
      return getProviderPrompt();
    })


    .then((providerAnswer) => {
      Object.assign(answers, providerAnswer);
      const provider = new Provider(answers[consts.provider]);
      return credentialsPrompts(provider)

        .then((keyAnswers) => {
          Object.assign(answers, keyAnswers);
          return machineConfigPrompts(provider);
        })

        .then((providerAnswers) => {
          Object.assign(answers, providerAnswers);
          return sshKeyPrompts(provider);
        })

        .then((sshKeyAnswers) => {
          Object.assign(answers, sshKeyAnswers);
          return machineCountPrompts();
        })

        .then((machineAnswers) => {
          Object.assign(answers, machineAnswers);
          return { provider, answers };
        });
    });
}

module.exports = { promptUser };
