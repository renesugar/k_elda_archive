const template = require('./packageTemplate.json');
const configureProvider = require('../configure-provider/package.json');
const baseInfrastructure = require('../base-infrastructure/package.json');

Object.keys(baseInfrastructure.dependencies).forEach((dep) => {
  if ((dep in configureProvider.dependencies)
    && (configureProvider.dependencies[dep] !== baseInfrastructure.dependencies[dep])) {
    throw new Error('configure-provider and base-infrastructure have different versions ' +
      `of the dependency '${dep}'.`);
  }
});

template.dependencies = Object.assign(
  configureProvider.dependencies, baseInfrastructure.dependencies);

console.log(JSON.stringify(template, null, 2));
