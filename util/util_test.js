const expect = require('chai').expect;
const path = require('path');
const util = require('./util');

describe('Create infrastructure path', function() {
  it('creates the right path', function() {
      let actual = util.infraPath('name');
      let expected = path.join(util.infraDirectory, 'name.js');
      expect(actual).to.equal(expected);
    });
});
