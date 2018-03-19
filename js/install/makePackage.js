const template = require('./packageTemplate.json');
const initializer = require('../initializer/package.json');
const baseInfrastructure = require('../base-infrastructure/package.json');

Object.keys(baseInfrastructure.dependencies).forEach((dep) => {
  if ((dep in initializer.dependencies)
    && (initializer.dependencies[dep] !== baseInfrastructure.dependencies[dep])) {
    throw new Error('initializer and base-infrastructure have different versions ' +
      `of the dependency '${dep}'.`);
  }
});

template.dependencies = Object.assign(
  initializer.dependencies, baseInfrastructure.dependencies);

console.log(JSON.stringify(template, null, 2));
