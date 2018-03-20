/* eslint-env mocha */
/* eslint-disable no-unused-expressions, no-underscore-dangle */
const expect = require('chai').expect;
const fs = require('fs');
const fsExtra = require('fs-extra');
const path = require('path');
const os = require('os');
const mock = require('mock-fs');
const rewire = require('rewire');
const sinon = require('sinon');

const consts = require('./constants');

const initializer = rewire('./initializer');
const p = rewire('./provider');

describe('configure-provider', () => {
  before(() => {
    initializer.__set__('log', sinon.stub());
  });

  describe('processAnswers()', () => {
    let revertConfig;

    beforeEach(() => {
      revertConfig = p.__set__('credentialsInfo', {
        provider: {
          credsTemplate: 'aTemplate',
          credsKeys: {
            key: 'key explanation',
            secret: 'secret explanation',
          },
          credsLocation: ['.some', 'file'],
        },
        providerNoTemplate: {
          credsLocation: ['.other', 'file'],
        },
        providerLongPath: {
          credsLocation: ['.a', 'long', 'path'],
        },
      });

      mock({ 'js/configure-provider/templates/aTemplate': '{{key}}-{{secret}}' });
    });

    afterEach(() => {
      mock.restore();
      revertConfig();
    });

    const credsSrc = path.join(os.homedir(), '.some', 'src');

    it('creates correct credentials from template', () => {
      const credsDst = path.join(os.homedir(), '.some', 'file');
      const provider = new p.Provider('provider');

      expect(fs.existsSync(credsDst)).to.be.false;

      initializer.processAnswers(provider, { key: 'myKey', secret: 'mySecret' });
      expect(fs.existsSync(credsDst)).to.be.true;
      expect(fs.readFileSync(credsDst, 'utf8')).to.equal('myKey-mySecret');
    });

    it('should create the directory path when needed', () => {
      const provider = new p.Provider('providerLongPath');

      const copyStub = sinon.stub();
      const mkdirpSpy = sinon.spy();
      const fsExtraMock = {
        copySync: copyStub,
        mkdirpSync: mkdirpSpy,
      };
      const revertFs = initializer.__set__('fsExtra', fsExtraMock);

      initializer.processAnswers(provider, {});

      expect(mkdirpSpy.calledOnce).to.be.true;
      expect(mkdirpSpy.getCall(0).args[0]).to.equal(path.join(os.homedir(), '.a', 'long'));

      revertFs();
    });

    it('creates correct credentials from path', () => {
      const credsDst = path.join(os.homedir(), '.other', 'file');
      const provider = new p.Provider('providerNoTemplate');

      const copySpy = sinon.spy();
      const fsExtraMock = {
        copySync: copySpy,
        mkdirpSync: fsExtra.mkdirpSync,
      };
      const revertFs = initializer.__set__('fsExtra', fsExtraMock);

      initializer.processAnswers(provider, { [consts.inputCredsPath]: credsSrc });

      expect(copySpy.calledOnce).to.be.true;
      expect(copySpy.getCall(0).args[0]).to.equal(credsSrc);
      expect(copySpy.getCall(0).args[1]).to.equal(credsDst);

      revertFs();
    });

    it('should not create credentials when they are not required', () => {
      const provider = new p.Provider('provider');

      const requiresCredsStub = sinon.stub(provider, 'requiresCreds');
      requiresCredsStub.returns(false);

      const getCredsPathSpy = sinon.spy(provider, 'getCredsPath');

      initializer.processAnswers(provider, { [consts.inputCredsPath]: credsSrc });

      expect(getCredsPathSpy.notCalled).to.be.true;

      getCredsPathSpy.restore();
      requiresCredsStub.restore();
    });
  });
});
