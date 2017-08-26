const template = require('./packageTemplate.json');
const initializer = require('../initializer/package.json');

template.dependencies = initializer.dependencies;
console.log(JSON.stringify(template, null, 2));
