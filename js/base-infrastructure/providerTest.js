/* eslint-env mocha */
/* eslint-disable no-unused-expressions, no-underscore-dangle */
const expect = require('chai').expect;
const mock = require('mock-fs');
const rewire = require('rewire');

const Provider = rewire('./provider');

describe('Provider', () => {
  let providerA;
  let providerB;

  before(() => {
    mock({
      'js/base-infrastructure/providers.json': `{
  "providerA": {
    "sizes": {
      "small": "size1",
      "medium": "size2",
      "large": "size3"
    },
    "regions": {
      "friendly1": "reg1",
      "friendly2": "reg2"
    },
    "hasPreemptible": true
  },
  "providerB": {
    "hasPreemptible": false
  }
}`,
    });

    providerA = new Provider('providerA');
    providerB = new Provider('providerB');
  });

  after(() => {
    mock.restore();
  });

  describe('name', () => {
    it('should have the right name', () => {
      expect(providerA.name).to.equal('providerA');
    });

    it('should get name for provider B', () => {
      expect(providerB.name).to.equal('providerB');
    });
  });

  describe('regions', () => {
    it('should have correct regions', () => {
      expect(providerA.regions).to.deep.equal({
        friendly1: 'reg1',
        friendly2: 'reg2',
      });
    });

    it('should not break when there are no regions', () => {
      expect(providerB.regions).to.deep.equal({});
    });
  });

  describe('sizes', () => {
    it('should have correct sizes', () => {
      expect(providerA.sizes).to.deep.equal({
        small: 'size1',
        medium: 'size2',
        large: 'size3',
      });
    });

    it('should not break when there are no sizes', () => {
      expect(providerB.sizes).to.deep.equal({});
    });
  });

  describe('hasPreemptible', () => {
    it('should have the right value of preemptible', () => {
      expect(providerA.hasPreemptible).to.be.true;
      expect(providerB.hasPreemptible).to.be.false;
    });
  });
});
