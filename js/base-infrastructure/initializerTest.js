/* eslint-env mocha */
/* eslint-disable no-unused-expressions, no-underscore-dangle */
const expect = require('chai').expect;
const fs = require('fs');
const rewire = require('rewire');
const sinon = require('sinon');

const consts = require('./constants');

const initializer = rewire('./initializer');

describe('base-infrastructure', () => {
  before(() => {
    initializer.__set__('log', sinon.stub());
  });

  describe('processAnswers()', () => {
    let revertFs;
    let revertFsExtra;
    let writeFileStub;

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

      revertFsExtra = initializer.__set__('fsExtra', fsExtraMock);
      revertFs = initializer.__set__('fs', fsMock);
    });

    afterEach(() => {
      revertFs();
      revertFsExtra();
      writeFileStub.resetHistory();
    });

    /**
     * Verifies that inputting the given answers to 'kelda base-infrastructure'
     * results in the given file describing the infrastructure.
     *
     * @param {Object} answers - Maps the name of a particular input to 'kelda
     *   base-infrastructure' to the answer provided for that input.
     * @param {string} expInfraFile - The file expected to be output by 'kelda
     *   base-infrastructure', when the given answers are provided as input.
     * @returns {void}
     */
    function checkInfrastructure(answers, expInfraFile) {
      initializer.processAnswers(answers);
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

      const expInfraFile = `/**
  * @param {Object} kelda - A Kelda module as returned by require('kelda').
  * @returns {kelda.Infrastructure} A Kelda infrastructure.
  */
function infraGetter(kelda) {
  const vmTemplate = new kelda.Machine({
    provider: 'provider',
    region: 'somewhere',
    size: 'big',
    preemptible: true,
  });

  return new kelda.Infrastructure({
    masters: vmTemplate.replicate(1),
    workers: vmTemplate.replicate(2),
  });
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

      const expInfraFile = `/**
  * @param {Object} kelda - A Kelda module as returned by require('kelda').
  * @returns {kelda.Infrastructure} A Kelda infrastructure.
  */
function infraGetter(kelda) {
  const vmTemplate = new kelda.Machine({
    provider: 'provider',
    region: 'somewhere',
    ram: 2,
    cpu: 1,
    preemptible: true,
  });

  return new kelda.Infrastructure({
    masters: vmTemplate.replicate(1),
    workers: vmTemplate.replicate(2),
  });
}

module.exports = infraGetter;
`;

      checkInfrastructure(answers, expInfraFile);
    });
    it('should not set the preemptible attribute when no value was given', () => {
      // The preemptible question will be skipped for providers that don't
      // support preemptible instances, so in those cases then preemptible entry
      // won't exist.
      const answers = {
        [consts.provider]: 'provider',
        [consts.size]: 'size',
        [consts.region]: 'somewhere',
        [consts.masterCount]: 1,
        [consts.workerCount]: 2,
      };

      const expInfraFile = `/**
  * @param {Object} kelda - A Kelda module as returned by require('kelda').
  * @returns {kelda.Infrastructure} A Kelda infrastructure.
  */
function infraGetter(kelda) {
  const vmTemplate = new kelda.Machine({
    provider: 'provider',
    region: 'somewhere',
    size: 'size',
  });

  return new kelda.Infrastructure({
    masters: vmTemplate.replicate(1),
    workers: vmTemplate.replicate(2),
  });
}

module.exports = infraGetter;
`;

      checkInfrastructure(answers, expInfraFile);
    });
  });
});
