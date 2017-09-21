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
const util = require('./init-util');

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
        'cli/command/init/templates/aTemplate': '{{key}}-{{secret}}',
        'cli/command/init/providers.json': `{
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

  describe('getSshKey()', () => {
    let revertFs;
    let readFileStub;
    let generateSshStub;
    let revertGenerator;

    before(() => {
      readFileStub = sinon.stub();
      readFileStub.returns('');
      const fsMock = {
        readFileSync: readFileStub,
      };
      revertFs = initializer.__set__('fs', fsMock);

      generateSshStub = sinon.stub();
      generateSshStub.returns(true);
      revertGenerator = initializer.__set__(
        'sshGenerateKeyPair', generateSshStub);
    });

    after(() => {
      revertFs();
      revertGenerator();
    });

    afterEach(() => {
      readFileStub.resetHistory();
      generateSshStub.resetHistory();
    });

    it('should not do anything, when the SSH key option is skipped', () => {
      /* eslint-disable no-undef */
      answers = { [consts.sshKeyOption]: consts.skip };
      initializer.getSshKey(answers);
      /* eslint-enable no-undef */
      expect(readFileStub.notCalled).to.be.true;
    });

    it('should generate a new SSH key when relevant', () => {
      initializer.getSshKey({ [consts.sshKeyOption]: consts.sshGenerateKey });
      expect(readFileStub.called).to.be.true;
      expect(generateSshStub.called).to.be.true;
    });

    it('should read an SSH key when given a path', () => {
      const keyPath = 'some/ssh/path';
      initializer.getSshKey({
        [consts.sshKeyOption]: consts.sshUseExistingKey,
        [consts.sshKeyPath]: keyPath,
      });

      expect(generateSshStub.notCalled).to.be.true;
      expect(readFileStub.called).to.be.true;
      expect(readFileStub.getCall(0).args[0]).to.equal(keyPath);
    });

    it('should throw an error when passed an unexpected SSH key option', () => {
      const expectFail = () => {
        initializer.getSshKey({ [consts.sshKeyOption]: 'badOption' });
      };

      expect(expectFail).to.throw();
    });
  });

  describe('processAnswers()', () => {
    let revertWriteCreds;
    let revertFs;
    let revertFsExtra;
    let revertSshKey;
    let writeCredsStub;
    let writeFileStub;
    let getSshKeyStub;
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

      getSshKeyStub = sinon.stub();
      writeCredsStub = sinon.stub();

      revertFsExtra = initializer.__set__('fsExtra', fsExtraMock);
      revertFs = initializer.__set__('fs', fsMock);
      revertSshKey = initializer.__set__('getSshKey', getSshKeyStub);
      revertWriteCreds = initializer.__set__(
        'writeProviderCreds', writeCredsStub);
    });

    afterEach(() => {
      revertWriteCreds();
      revertFs();
      revertFsExtra();
      revertSshKey();
      writeFileStub.resetHistory();
      writeCredsStub.resetHistory();
      getSshKeyStub.resetBehavior();
    });

    function checkInfrastructure(answers, expInfraFile) {
      initializer.processAnswers(provider, answers);
      expect(getSshKeyStub.called).to.be.true;
      expect(writeCredsStub.getCall(0).args[0]).to.equal(provider);
      expect(writeCredsStub.getCall(0).args[1]).to.equal(answers);
      expect(writeFileStub.getCall(0).args[0]).to.equal(
        util.infraPath(answers.name));
      expect(writeFileStub.getCall(0).args[1]).to.equal(expInfraFile);
    }

    it('should set SSH key and size when provided', () => {
      getSshKeyStub.returns('key');
      const answers = {
        [consts.name]: 'foo',
        [consts.provider]: 'provider',
        [consts.size]: 'big',
        // Size overrides RAM and CPU.
        [consts.ram]: 2,
        [consts.cpu]: 1,
        [consts.preemptible]: true,
        [consts.region]: 'somewhere',
        [consts.sshKey]: 'keyOption',
        [consts.masterCount]: 1,
        [consts.workerCount]: 2,
      };

      const expInfraFile = `function infraGetter(quilt) {
  const inf = quilt.createDeployment({namespace: 'quilt-deployment'});

  var vmTemplate = new quilt.Machine({
    provider: 'provider',
    region: 'somewhere',
    size: 'big',
    sshKeys: ['key'],
    preemptible: true
  });

  inf.deploy(vmTemplate.asMaster().replicate(1));
  inf.deploy(vmTemplate.asWorker().replicate(2));

  return inf;
}

module.exports = infraGetter;
`;
      checkInfrastructure(answers, expInfraFile);
    });
    it('should omit SSH key and size when not provided and use RAM/CPU ' +
      'instead', () => {
      getSshKeyStub.returns('');
      const answers = {
        [consts.name]: 'foo',
        [consts.provider]: 'provider',
        [consts.ram]: 2,
        [consts.cpu]: 1,
        [consts.preemptible]: true,
        [consts.region]: 'somewhere',
        [consts.sshKey]: 'noKey',
        [consts.masterCount]: 1,
        [consts.workerCount]: 2,
      };

      const expInfraFile = `function infraGetter(quilt) {
  const inf = quilt.createDeployment({namespace: 'quilt-deployment'});

  var vmTemplate = new quilt.Machine({
    provider: 'provider',
    region: 'somewhere',
    ram: 2,
    cpu: 1,
    preemptible: true
  });

  inf.deploy(vmTemplate.asMaster().replicate(1));
  inf.deploy(vmTemplate.asWorker().replicate(2));

  return inf;
}

module.exports = infraGetter;
`;

      checkInfrastructure(answers, expInfraFile);
    });
  });
});
