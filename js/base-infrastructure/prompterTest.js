/* eslint-env mocha */
const expect = require('chai').expect;
const prompter = require('./prompter');

describe('Prompter', () => {
  describe('allProviders()', () => {
    it('should return all supported providers', () => {
      expect(prompter.allProviders()).to.include.members(
        ['Vagrant', 'Amazon', 'Google', 'DigitalOcean']);
      expect(prompter.allProviders()).to.have.lengthOf(4);
    });
  });
  describe('isNumber()', () => {
    it('should not error when passed an integer', () => {
      expect(() => prompter.isNumber('8')).to.not.throw();
      expect(() => prompter.isNumber('108')).to.not.throw();
    });
    it('should error when passed a non-integer number', () => {
      expect(() => prompter.isNumber('1.8')).to.throw();
      expect(() => prompter.isNumber('0.8')).to.throw();
    });
    it('should error when passed a negative number', () => {
      expect(() => prompter.isNumber('-8')).to.throw();
      expect(() => prompter.isNumber('-0.8')).to.throw();
    });
    it('should error when passed a non-number', () => {
      expect(() => prompter.isNumber('hello')).to.throw();
      expect(() => prompter.isNumber('8hello')).to.throw();
    });
  });
  describe('getInquirerDescriptions()', () => {
    it('should return the correct map', () => {
      const actualDescriptions = prompter.getInquirerDescriptions(
        { smallName: 'smallSize', mediumName: 'mediumSize' });
      expect(actualDescriptions).to.include.deep.members([
        { name: 'smallName (smallSize)', value: 'smallSize' },
        { name: 'mediumName (mediumSize)', value: 'mediumSize' },
      ]);
      expect(actualDescriptions).to.have.lengthOf(2);
    });
    it('should not error when given an empty object', () => {
      expect(() => prompter.getInquirerDescriptions({})).to.not.throw();
      expect(prompter.getInquirerDescriptions({})).to.have.lengthOf(0);
    });
  });
});
