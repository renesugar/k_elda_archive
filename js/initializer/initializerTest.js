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
const Provider = rewire('./provider');

describe('Initializer', () => {
  before(() => {
    initializer.__set__('log', sinon.stub());
  });

  describe('writeProviderCreds()', () => {
    let revertConfig;

    beforeEach(() => {
      revertConfig = Provider.__set__('providerConfig', {
        provider: {
          credsTemplate: 'aTemplate',
          credsKeys: {
            key: 'key explanation',
            secret: 'secret explanation',
          },
        },
      });

      mock({
        'js/initializer/templates/aTemplate': '{{key}}-{{secret}}',
        'js/initializer/providers.json': `{
    "provider": {
      "credsLocation": [".some", "file"]
    }
  }`,
      });
    });

    afterEach(() => {
      mock.restore();
      revertConfig();
    });

    const credsSrc = path.join(os.homedir(), '.some', 'src');
    const credsDst = path.join(os.homedir(), '.some', 'file');

    it('creates correct credentials from template', () => {
      const provider = new Provider('provider');

      const credsDest = path.join(os.homedir(), '.some', 'file');
      expect(fs.existsSync(credsDest)).to.be.false;

      fs.mkdir(path.join(os.homedir(), '.some'));
      initializer.writeProviderCreds(
        provider,
        { key: 'myKey', secret: 'mySecret' });
      expect(fs.existsSync(credsDest)).to.be.true;

      expect(fs.readFileSync(credsDest, 'utf8')).to.equal('myKey-mySecret');
    });

    it('creates correct credentials from path', () => {
      const provider = new Provider('provider');

      const copySpy = sinon.spy();
      const fsExtraMock = {
        copySync: copySpy,
        mkdirpSync: fsExtra.mkdirpSync,
      };
      const revertFs = initializer.__set__('fsExtra', fsExtraMock);

      // The template will be undefined when the provider requires a
      // credentials path instead of template input.
      provider.credsTemplate = undefined;

      initializer.writeProviderCreds(
        provider,
        { [consts.inputCredsPath]: credsSrc });

      expect(copySpy.calledOnce).to.be.true;
      expect(copySpy.getCall(0).args[0]).to.equal(credsSrc);
      expect(copySpy.getCall(0).args[1]).to.equal(credsDst);

      revertFs();
    });

    it('should not create credentials when they are not required', () => {
      const provider = new Provider('provider');

      const requiresCredsStub = sinon.stub(provider, 'requiresCreds');
      requiresCredsStub.returns(false);

      const getCredsPathSpy = sinon.spy(provider, 'getCredsPath');

      initializer.writeProviderCreds(
        provider,
        { [consts.inputCredsPath]: credsSrc });

      expect(getCredsPathSpy.notCalled).to.be.true;

      getCredsPathSpy.restore();
      requiresCredsStub.restore();
    });
  });

  describe('processAnswers()', () => {
    let revertWriteCreds;
    let revertFs;
    let revertFsExtra;
    let writeCredsStub;
    let writeFileStub;
    const provider = { provider: 'provider' };

    beforeEach(() => {
      const fsExtraMock = {
        mkdirpSync: sinon.stub(),
      };

      writeFileStub = sinon.stub();
      const fsMock = {
        existsSync: sinon.stub().returns(true),
        writeFileSync: writeFileStub,
        readFileSync: fs.readFileSync,
      };

      writeCredsStub = sinon.stub();

      revertFsExtra = initializer.__set__('fsExtra', fsExtraMock);
      revertFs = initializer.__set__('fs', fsMock);
      revertWriteCreds = initializer.__set__(
        'writeProviderCreds', writeCredsStub);
    });

    afterEach(() => {
      revertWriteCreds();
      revertFs();
      revertFsExtra();
      writeFileStub.resetHistory();
      writeCredsStub.resetHistory();
    });

    /**
     * Verifies that inputting the given answers to 'kelda init' results
     * in the given file describing the infrastructure.
     *
     * @param {Object} answers - Maps the name of a particular input to 'kelda init'
     *   to the answer provided for that input.
     * @param {string} expInfraFile - The file expected to be output by 'kelda init',
     *   when the given answers are provided as input.
     * @returns {void}
     */
    function checkInfrastructure(answers, expInfraFile) {
      initializer.processAnswers(provider, answers);
      expect(writeCredsStub.getCall(0).args[0]).to.equal(provider);
      expect(writeCredsStub.getCall(0).args[1]).to.equal(answers);
      expect(writeFileStub.getCall(0).args[0]).to.equal(consts.baseInfraLocation);
      expect(writeFileStub.getCall(0).args[1]).to.equal(expInfraFile);
    }

    it('should set SSH key and size when provided', () => {
      const answers = {
        [consts.provider]: 'provider',
        [consts.size]: 'big',
        // Size overrides RAM and CPU.
        [consts.ram]: 2,
        [consts.cpu]: 1,
        [consts.preemptible]: true,
        [consts.region]: 'somewhere',
        [consts.masterCount]: 1,
        [consts.workerCount]: 2,
      };

      const expInfraFile = `function infraGetter(kelda) {

  var vmTemplate = new kelda.Machine({
    provider: 'provider',
    region: 'somewhere',
    size: 'big',
    preemptible: true
  });

  return new kelda.Infrastructure(
    vmTemplate.replicate(1),
    vmTemplate.replicate(2));
}

module.exports = infraGetter;
`;
      checkInfrastructure(answers, expInfraFile);
    });
    it('should omit SSH key and size when not provided and use RAM/CPU ' +
      'instead', () => {
      const answers = {
        [consts.provider]: 'provider',
        [consts.ram]: 2,
        [consts.cpu]: 1,
        [consts.preemptible]: true,
        [consts.region]: 'somewhere',
        [consts.masterCount]: 1,
        [consts.workerCount]: 2,
      };

      const expInfraFile = `function infraGetter(kelda) {

  var vmTemplate = new kelda.Machine({
    provider: 'provider',
    region: 'somewhere',
    ram: 2,
    cpu: 1,
    preemptible: true
  });

  return new kelda.Infrastructure(
    vmTemplate.replicate(1),
    vmTemplate.replicate(2));
}

module.exports = infraGetter;
`;

      checkInfrastructure(answers, expInfraFile);
    });
  });
});
