/* eslint-env mocha */
/* eslint-disable no-unused-expressions, no-underscore-dangle */
const expect = require('chai').expect;
const os = require('os');
const path = require('path');
const mock = require('mock-fs');
const rewire = require('rewire');

const p = rewire('./provider');

describe('Provider', () => {
  let revertConfig;
  const credentialsPathA = path.join(os.homedir(), '.some/file');
  const credentialsPathC = path.join(os.homedir(), '.another/file');

  before(() => {
    revertConfig = p.__set__('credentialsInfo', {
      providerA: {
        credsTemplate: 'aTemplate',
        credsKeys: {
          key: 'aKey',
          secret: 'aSecret',
        },
        credsLocation: ['.some', 'file'],
      },
      providerB: {},
      providerC: {
        credsLocation: ['.another', 'file'],
      },
    });
    mock({ [credentialsPathC]: 'I do exist' });
  });

  after(() => {
    mock.restore();
    revertConfig();
  });

  let providerA;
  let providerB;
  let providerC;

  describe('getName()', () => {
    it('should get name for provider A', () => {
      providerA = new p.Provider('providerA');
      expect(providerA.getName()).to.equal('providerA');
    });

    it('should get name for provider B', () => {
      providerB = new p.Provider('providerB');
      expect(providerB.getName()).to.equal('providerB');
    });
  });

  describe('getCredsKeys()', () => {
    it('should get the keys', () => {
      expect(providerA.getCredsKeys()).to.deep.equal({
        key: 'aKey',
        secret: 'aSecret',
      });
    });

    it('should not break when there are no keys', () => {
      expect(providerB.getCredsKeys()).to.deep.equal({});
    });
  });

  describe('getCredsTemplate()', () => {
    it('should return the name of the template file', () => {
      expect(providerA.getCredsTemplate()).to.equal('aTemplate');
    });

    it('should not break when there is no template', () => {
      expect(providerB.getCredsTemplate()).to.equal('');
    });
  });

  describe('requiresCreds()', () => {
    it('should return true if the provider requires credentials', () => {
      expect(providerA.requiresCreds()).to.be.true;
    });

    it('should return false if the provider does not require credentials',
      () => {
        expect(providerB.requiresCreds()).to.be.false;
      });
  });

  describe('credsExist()', () => {
    it('should return false if there are no existing credentials', () => {
      expect(providerA.credsExist()).to.be.false;
    });

    it('should return true if credentials exist', () => {
      providerC = new p.Provider('providerC');
      // This file is mocked in the `before` hook.
      expect(providerC.credsExist()).to.be.true;
    });
  });

  describe('getCredsPath()', () => {
    it('should return the correct path', () => {
      expect(providerC.getCredsPath()).to.equal(credentialsPathC);
    });

    it('should return correct path, also when the file does not exist',
      () => {
        expect(providerA.getCredsPath()).to.equal(credentialsPathA);
      });
  });

  describe('allProviders', () => {
    it('should return the right provider names', () => {
      expect(p.allProviders()).to.include.members(['providerA', 'providerB', 'providerC']);
      expect(p.allProviders()).to.have.lengthOf(3);
    });
  });
});
