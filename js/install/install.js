const os = require('os');
const fs = require('fs');

const platform = os.platform();
switch (platform) {
  case 'darwin':
    fs.symlinkSync('./kelda_darwin', './kelda');
    break;
  case 'linux':
    fs.symlinkSync('./kelda_linux', './kelda');
    break;
  default:
    throw new Error(`Unrecognized operating system ${platform}. Bailing.`);
}
