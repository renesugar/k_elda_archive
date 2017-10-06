/* eslint-env mocha */

/* eslint-disable import/no-extraneous-dependencies, no-underscore-dangle */
const chai = require('chai');
const chaiSubset = require('chai-subset');
const rewire = require('rewire');
const sinon = require('sinon');

const b = rewire('./bindings.js');

chai.use(chaiSubset);
const { expect } = chai;

describe('Bindings', () => {
  let deployment;
  beforeEach(() => {
    b.resetGlobals();
    deployment = new b.Deployment();
  });

  const checkMachines = function checkMachines(expected) {
    const { machines } = deployment.toQuiltRepresentation();
    expect(machines).to.have.lengthOf(expected.length)
      .and.containSubset(expected);
  };

  const checkContainers = function checkContainers(expected) {
    const { containers } = deployment.toQuiltRepresentation();
    expect(containers).to.have.lengthOf(expected.length)
      .and.containSubset(expected);
  };

  const checkPlacements = function checkPlacements(expected) {
    const { placements } = deployment.toQuiltRepresentation();
    expect(placements).to.have.lengthOf(expected.length)
      .and.containSubset(expected);
  };

  const checkLoadBalancers = function checkLoadBalancers(expected) {
    const { loadBalancers } = deployment.toQuiltRepresentation();
    expect(loadBalancers).to.have.lengthOf(expected.length)
      .and.containSubset(expected);
  };

  const checkConnections = function checkConnections(expected) {
    const { connections } = deployment.toQuiltRepresentation();
    expect(connections).to.have.lengthOf(expected.length)
      .and.containSubset(expected);
  };

  describe('Machine', () => {
    it('basic', () => {
      deployment.deploy([new b.Machine({
        role: 'Worker',
        provider: 'Amazon',
        region: 'us-west-2',
        size: 'm4.large',
        cpu: new b.Range(2, 4),
        ram: new b.Range(4, 8),
        diskSize: 32,
        sshKeys: ['key1', 'key2'],
      })]);
      checkMachines([{
        id: 'ae657514e0aa41ed95d9e27c3f3c9b2ff23bd05e',
        role: 'Worker',
        provider: 'Amazon',
        region: 'us-west-2',
        size: 'm4.large',
        cpu: new b.Range(2, 4),
        ram: new b.Range(4, 8),
        diskSize: 32,
        sshKeys: ['key1', 'key2'],
      }]);
    });
    it('errors when passed invalid optional arguments', () => {
      expect(() => new b.Machine({ badArg: 'foo' })).to
        .throw('Unrecognized keys passed to Machine constructor: badArg');
      expect(() => new b.Machine({
        badArg: 'foo', provider: 'Amazon', alsoBad: 'bar' }))
        .to.throw('Unrecognized keys passed to Machine constructor: ');
    });
    it('hash independent of SSH keys', () => {
      deployment.deploy([new b.Machine({
        role: 'Worker',
        provider: 'Amazon',
        region: 'us-west-2',
        size: 'm4.large',
        cpu: new b.Range(2, 4),
        ram: new b.Range(4, 8),
        diskSize: 32,
        sshKeys: ['key3'],
      })]);
      checkMachines([{
        id: 'ae657514e0aa41ed95d9e27c3f3c9b2ff23bd05e',
        role: 'Worker',
        provider: 'Amazon',
        region: 'us-west-2',
        size: 'm4.large',
        cpu: new b.Range(2, 4),
        ram: new b.Range(4, 8),
        diskSize: 32,
        sshKeys: ['key3'],
      }]);
    });
    it('replicate', () => {
      const baseMachine = new b.Machine({ provider: 'Amazon' });
      deployment.deploy(baseMachine.asMaster().replicate(2));
      checkMachines([
        {
          id: '38f289007e41382ce4e2773508609674bac7df52',
          role: 'Master',
          provider: 'Amazon',
        },
        {
          id: 'e23719b2160e4b42c6bbca72567220833fac68da',
          role: 'Master',
          provider: 'Amazon',
        },
      ]);
    });
    it('replicate independent', () => {
      const baseMachine = new b.Machine({ provider: 'Amazon' });
      const machines = baseMachine.asMaster().replicate(2);
      machines[0].sshKeys.push('key');
      deployment.deploy(machines);
      checkMachines([
        {
          id: '38f289007e41382ce4e2773508609674bac7df52',
          role: 'Master',
          provider: 'Amazon',
          sshKeys: ['key'],
        },
        {
          id: 'e23719b2160e4b42c6bbca72567220833fac68da',
          role: 'Master',
          provider: 'Amazon',
        },
      ]);
    });
    it('set floating IP', () => {
      const baseMachine = new b.Machine({
        provider: 'Amazon',
        floatingIp: 'xxx.xxx.xxx.xxx',
      });
      deployment.deploy(baseMachine.asMaster());
      checkMachines([{
        id: 'bc2c5392f98b605e90007056e580a42c0c3f960d',
        role: 'Master',
        provider: 'Amazon',
        floatingIp: 'xxx.xxx.xxx.xxx',
        sshKeys: [],
      }]);
    });
    it('preemptible attribute', () => {
      deployment.deploy(new b.Machine({
        provider: 'Amazon',
        preemptible: true,
      }).asMaster());
      checkMachines([{
        id: '893cfbfaccf6aa6e518f1757dadb07ffb936082f',
        role: 'Master',
        provider: 'Amazon',
        preemptible: true,
      }]);
    });
  });

  describe('Container', () => {
    it('basic', () => {
      const container = new b.Container('host', 'image');
      container.deploy(deployment);
      checkContainers([{
        id: '293fc7ad8a799d3cf2619a3db7124b0459f395cb',
        image: new b.Image('image'),
        hostname: 'host',
        command: [],
        env: {},
        filepathToContent: {},
      }]);
    });
    it('errors when passed invalid optional arguments', () => {
      expect(() => new b.Container('host', 'image', { badArg: 'foo' })).to
        .throw('Unrecognized keys passed to Container constructor: badArg');
      expect(() => new b.Container('host', 'image',
        { badArg: 'foo', command: [], alsoBad: 'bar' }))
        .to.throw('Unrecognized keys passed to Container constructor: ');
    });
    it('containers are not duplicated', () => {
      const container = new b.Container('host', 'image');
      container.deploy(deployment);
      container.deploy(deployment);
      checkContainers([{
        id: '293fc7ad8a799d3cf2619a3db7124b0459f395cb',
        image: new b.Image('image'),
        hostname: 'host',
        command: [],
        env: {},
        filepathToContent: {},
      }]);
    });
    it('command', () => {
      const container = new b.Container('host', 'image', {
        command: ['arg1', 'arg2'],
      });
      container.deploy(deployment);
      checkContainers([{
        id: '9d0d496d49bed06e7c76c2b536d7520ccc1717f2',
        image: new b.Image('image'),
        command: ['arg1', 'arg2'],
        hostname: 'host',
        env: {},
        filepathToContent: {},
      }]);
    });
    it('env', () => {
      const c = new b.Container('host', 'image');
      c.env.foo = 'bar';
      c.deploy(deployment);
      checkContainers([{
        id: '299619d3fb4b89fd5cc822983bc3fbcced2f0a98',
        image: new b.Image('image'),
        command: [],
        env: { foo: 'bar' },
        hostname: 'host',
        filepathToContent: {},
      }]);
    });
    it('hostname', () => {
      const c = new b.Container('host', new b.Image('image'));
      c.deploy(deployment);
      expect(c.getHostname()).to.equal('host.q');
      checkContainers([{
        id: '293fc7ad8a799d3cf2619a3db7124b0459f395cb',
        image: new b.Image('image'),
        command: [],
        env: {},
        filepathToContent: {},
        hostname: 'host',
      }]);
    });
    it('repeated hostnames don\'t conflict', () => {
      const x = new b.Container('host', 'image');
      const y = new b.Container('host', 'image');
      x.deploy(deployment);
      y.deploy(deployment);
      checkContainers([{
        id: '293fc7ad8a799d3cf2619a3db7124b0459f395cb',
        image: new b.Image('image'),
        command: [],
        env: {},
        filepathToContent: {},
        hostname: 'host',
      }, {
        id: '968bcf8c6d235afbc88aec8e1bdddc506714a0b8',
        image: new b.Image('image'),
        command: [],
        env: {},
        filepathToContent: {},
        hostname: 'host2',
      }]);
    });
    it('Container.hostname and LoadBalancer.hostname don\'t conflict', () => {
      const container = new b.Container('foo', 'image');
      const serv = new b.LoadBalancer('foo', []);
      expect(container.getHostname()).to.equal('foo.q');
      expect(serv.hostname()).to.equal('foo2.q');
    });
    it('hostnames returned by uniqueHostname cannot be reused', () => {
      const containerA = new b.Container('host', 'ignoreme');
      const containerB = new b.Container('host', 'ignoreme');
      const containerC = new b.Container('host2', 'ignoreme');
      const hostnames = [containerA, containerB, containerC]
        .map(c => c.getHostname());
      const hostnameSet = new Set(hostnames);
      expect(hostnames.length).to.equal(hostnameSet.size);
    });
    it('clone increments existing index if one exists', () => {
      const containerA = new b.Container('host', 'ignoreme');
      const containerB = containerA.clone();
      const containerC = containerB.clone();
      expect(containerA.getHostname()).to.equal('host.q');
      expect(containerB.getHostname()).to.equal('host2.q');
      expect(containerC.getHostname()).to.equal('host3.q');
    });
    it('duplicate hostname causes error', () => {
      const x = new b.Container('host', 'image');
      x.hostname = 'host';
      x.deploy(deployment);
      const y = new b.Container('host', 'image');
      y.hostname = 'host';
      y.deploy(deployment);
      expect(() => deployment.toQuiltRepresentation()).to
        .throw('hostname "host" used multiple times');
    });
    it('image dockerfile', () => {
      const z = new b.Container('host', new b.Image('name', 'dockerfile'));
      z.deploy(deployment);
      checkContainers([{
        id: 'fbc9aedb5af0039b8cf09bca2ef5771467b44085',
        image: new b.Image('name', 'dockerfile'),
        hostname: 'host',
        command: [],
        env: {},
        filepathToContent: {},
      }]);
    });
  });

  describe('Container attributes', () => {
    const hostname = 'host';
    const image = new b.Image('image');
    const command = ['arg1', 'arg2'];
    const env = { foo: 'bar' };
    const filepathToContent = { qux: 'quuz' };
    it('with*', () => {
      // The blueprint ID is different than the Container created with the
      // constructor because the hostname ID increases with each with*
      // call.
      const id = 'f5c3e0fa3843e6fa149289d476f507831a45654d';
      const container = new b.Container(hostname, image, {
        command,
      }).withEnv(env)
        .withFiles(filepathToContent);
      container.deploy(deployment);
      checkContainers([{
        id,
        image,
        command,
        env,
        filepathToContent,
        hostname: 'host3',
      }]);
    });
    it('constructor', () => {
      const id = '9f9d0c0868163eda5b4ad5861c9558f055508959';
      const container = new b.Container(hostname, image, {
        command, env, filepathToContent,
      });
      container.deploy(deployment);
      checkContainers([{
        id, hostname, image, command, env, filepathToContent,
      }]);
    });
  });

  describe('Placement', () => {
    let target;
    beforeEach(() => {
      target = new b.Container('host', 'image');
      target.deploy(deployment);
    });
    it('MachineRule size, region, provider', () => {
      target.placeOn({
        size: 'm4.large',
        region: 'us-west-2',
        provider: 'Amazon',
      });
      checkPlacements([{
        targetContainerID: '293fc7ad8a799d3cf2619a3db7124b0459f395cb',
        exclusive: false,
        region: 'us-west-2',
        provider: 'Amazon',
        size: 'm4.large',
      }]);
    });
    it('MachineRule size, provider', () => {
      target.placeOn({
        size: 'm4.large',
        provider: 'Amazon',
      });
      checkPlacements([{
        targetContainerID: '293fc7ad8a799d3cf2619a3db7124b0459f395cb',
        exclusive: false,
        provider: 'Amazon',
        size: 'm4.large',
      }]);
    });
    it('MachineRule floatingIp', () => {
      target.placeOn({
        floatingIp: 'xxx.xxx.xxx.xxx',
      });
      checkPlacements([{
        targetContainerID: '293fc7ad8a799d3cf2619a3db7124b0459f395cb',
        exclusive: false,
        floatingIp: 'xxx.xxx.xxx.xxx',
      }]);
    });
  });
  describe('LoadBalancer', () => {
    it('basic', () => {
      const lb = new b.LoadBalancer('web_tier', [new b.Container('host', 'nginx')]);
      lb.deploy(deployment);
      checkLoadBalancers([{
        name: 'web_tier',
        hostnames: ['host'],
      }]);
    });
    it('multiple containers', () => {
      const lb = new b.LoadBalancer('web_tier', [
        new b.Container('host', 'nginx'),
        new b.Container('host', 'nginx'),
      ]);
      lb.deploy(deployment);
      checkLoadBalancers([{
        name: 'web_tier',
        hostnames: [
          'host',
          'host2',
        ],
      }]);
    });
    it('duplicate load balancers', () => {
      /* Conflicting load balancer names.  We need to generate a couple of dummy
               containers so that the two deployed containers have _refID's
               that are sorted differently lexicographically and numerically. */
      for (let i = 0; i < 2; i += 1) {
        new b.Container('host', 'image'); // eslint-disable-line no-new
      }
      const lb = new b.LoadBalancer('foo', [new b.Container('host', 'image')]);
      lb.deploy(deployment);
      for (let i = 0; i < 7; i += 1) {
        new b.Container('host', 'image'); // eslint-disable-line no-new
      }
      const lb2 = new b.LoadBalancer('foo', [new b.Container('host', 'image')]);
      lb2.deploy(deployment);
      checkLoadBalancers([
        {
          name: 'foo',
          hostnames: ['host3'],
        },
        {
          name: 'foo2',
          hostnames: ['host11'],
        },
      ]);
    });
    it('get LoadBalancer hostname', () => {
      const foo = new b.LoadBalancer('foo', []);
      expect(foo.hostname()).to.equal('foo.q');
    });
  });
  describe('AllowFrom', () => {
    let foo;
    let bar;
    let fooLoadBalancer;
    beforeEach(() => {
      foo = new b.Container('foo', 'image');
      foo.deploy(deployment);
      fooLoadBalancer = new b.LoadBalancer('fooLoadBalancer', [foo]);
      bar = new b.Container('bar', 'image');
      bar.deploy(deployment);
      fooLoadBalancer.deploy(deployment);
    });
    it('autobox port ranges', () => {
      bar.allowFrom(foo, 80);
      checkConnections([{
        from: 'foo',
        to: 'bar',
        minPort: 80,
        maxPort: 80,
      }]);
    });
    it('port', () => {
      bar.allowFrom(foo, new b.Port(80));
      checkConnections([{
        from: 'foo',
        to: 'bar',
        minPort: 80,
        maxPort: 80,
      }]);
    });
    it('port range', () => {
      bar.allowFrom(foo, new b.PortRange(80, 85));
      checkConnections([{
        from: 'foo',
        to: 'bar',
        minPort: 80,
        maxPort: 85,
      }]);
    });
    it('connect to invalid port range', () => {
      expect(() => foo.allowFrom(bar, true)).to
        .throw('Input argument must be a number or a Range');
    });
    it('allow connections to publicInternet', () => {
      b.publicInternet.allowFrom(foo, 80);
      checkConnections([{
        from: 'foo',
        to: 'public',
        minPort: 80,
        maxPort: 80,
      }]);
    });
    it('allow connections from publicInternet', () => {
      foo.allowFrom(b.publicInternet, 80);
      checkConnections([{
        from: 'public',
        to: 'foo',
        minPort: 80,
        maxPort: 80,
      }]);
    });
    it('connect to LoadBalancer', () => {
      fooLoadBalancer.allowFrom(bar, 80);
      checkConnections([{
        from: 'bar',
        to: 'fooLoadBalancer',
        minPort: 80,
        maxPort: 80,
      }]);
    });
    it('connect to publicInternet port range', () => {
      expect(() =>
        b.publicInternet.allowFrom(foo, new b.PortRange(80, 81))).to
        .throw('public internet can only connect to single ports ' +
                        'and not to port ranges');
    });
    it('connect from publicInternet port range', () => {
      expect(() =>
        foo.allowFrom(b.publicInternet, new b.PortRange(80, 81))).to
        .throw('public internet can only connect to single ports ' +
                        'and not to port ranges');
    });
    it('allowFrom non-container', () => {
      expect(() => foo.allowFrom(10, 10)).to
        .throw('Containers can only connect to other containers. ' +
                    'Check that you\'re allowing connections from a ' +
                    'container or list of containers, and not from a LoadBalancer ' +
                    'or other object.');
    });
  });
  describe('allow', () => {
    let foo;
    let bar;
    let qux;
    let quuz;
    let fooBarGroup;
    let quxQuuzGroup;
    let lb;
    beforeEach(() => {
      foo = new b.Container('foo', 'image');
      bar = new b.Container('bar', 'image');
      qux = new b.Container('qux', 'image');
      quuz = new b.Container('quuz', 'image');

      fooBarGroup = [foo, bar];
      quxQuuzGroup = [qux, quuz];
      lb = new b.LoadBalancer('serv', [foo, bar, qux, quuz]);

      foo.deploy(deployment);
      bar.deploy(deployment);
      qux.deploy(deployment);
      quuz.deploy(deployment);
      lb.deploy(deployment);
    });

    it('both src and dst are lists', () => {
      b.allow(quxQuuzGroup, fooBarGroup, 80);
      checkConnections([
        { from: 'qux', to: 'foo', minPort: 80, maxPort: 80 },
        { from: 'qux', to: 'bar', minPort: 80, maxPort: 80 },
        { from: 'quuz', to: 'foo', minPort: 80, maxPort: 80 },
        { from: 'quuz', to: 'bar', minPort: 80, maxPort: 80 },
      ]);
    });

    it('dst is a list', () => {
      b.allow(qux, fooBarGroup, 80);
      checkConnections([
        { from: 'qux', to: 'foo', minPort: 80, maxPort: 80 },
        { from: 'qux', to: 'bar', minPort: 80, maxPort: 80 },
      ]);
    });

    it('src is a list', () => {
      b.allow(fooBarGroup, qux, 80);
      checkConnections([
        { from: 'foo', to: 'qux', minPort: 80, maxPort: 80 },
        { from: 'bar', to: 'qux', minPort: 80, maxPort: 80 },
      ]);
    });

    it('src is public', () => {
      b.allow(b.publicInternet, fooBarGroup, 80);
      checkConnections([
        { from: 'public', to: 'foo', minPort: 80, maxPort: 80 },
        { from: 'public', to: 'bar', minPort: 80, maxPort: 80 },
      ]);
    });

    it('dst is public', () => {
      b.allow(fooBarGroup, b.publicInternet, 80);
      checkConnections([
        { from: 'foo', to: 'public', minPort: 80, maxPort: 80 },
        { from: 'bar', to: 'public', minPort: 80, maxPort: 80 },
      ]);
    });

    it('dst is a LoadBalancer', () => {
      b.allow(fooBarGroup, lb, 80);
      checkConnections([
        { from: 'foo', to: 'serv', minPort: 80, maxPort: 80 },
        { from: 'bar', to: 'serv', minPort: 80, maxPort: 80 },
      ]);
    });
  });
  describe('Vet', () => {
    let foo;
    const deploy = () => deployment.toQuiltRepresentation();
    beforeEach(() => {
      foo = new b.LoadBalancer('foo', []);
      foo.deploy(deployment);
    });
    it('should error when given namespace contains upper case letters', () => {
      deployment = new b.Deployment({ namespace: 'BadNamespace' });
      expect(deploy).to.throw('namespace "BadNamespace" contains ' +
                  'uppercase letters. Namespaces must be lowercase.');
    });
    it('connect from undeployed container', () => {
      foo.allowFrom(new b.Container('baz', 'image'), 80);
      expect(deploy).to.throw(
        'connection {"from":"baz","maxPort":80,"minPort":80,' +
                '"to":"foo"} references an undefined hostname: baz');
    });
    it('duplicate image', () => {
      (new b.Container('host', new b.Image('img', 'dk'))).deploy(deployment);
      (new b.Container('host', new b.Image('img', 'dk'))).deploy(deployment);
      expect(deploy).to.not.throw();
    });
    it('duplicate image with different Dockerfiles', () => {
      (new b.Container('host', new b.Image('img', 'dk'))).deploy(deployment);
      (new b.Container('host', new b.Image('img', 'dk2'))).deploy(deployment);
      expect(deploy).to.throw('img has differing Dockerfiles');
    });
    it('machines with same regions/providers', () => {
      deployment.deploy([new b.Machine({
        provider: 'Amazon',
        region: 'us-west-2',
      }), new b.Machine({
        provider: 'Amazon',
        region: 'us-west-2',
      })]);
      expect(deploy).to.not.throw();
    });
    it('machines with different regions', () => {
      deployment.deploy([new b.Machine({
        provider: 'Amazon',
        region: 'us-west-2',
      }), new b.Machine({
        provider: 'Amazon',
        region: 'us-east-2',
      })]);
      expect(deploy).to.throw('All machines must have the same provider and region. '
        + 'Found providers \'Amazon\' in region \'us-west-2\' and \'Amazon\' in '
        + 'region \'us-east-2\'.');
    });
    it('machines with different providers', () => {
      deployment.deploy([new b.Machine({
        provider: 'Amazon',
        region: '',
      }), new b.Machine({
        provider: 'DigitalOcean',
        region: '',
      })]);
      expect(deploy).to.throw('All machines must have the same provider and region. '
        + 'Found providers \'Amazon\' in region \'\' and \'DigitalOcean\' in '
        + 'region \'\'.');
    });
  });
  describe('Deployment', () => {
    it('no args', () => {
      expect(() => new b.Deployment()).to.not.throw();
    });
    it('should error when given invalid arguments', () => {
      expect(() => new b.Deployment({ badArg: 'foo' }))
        .to.throw('Unrecognized keys passed to Deployment constructor: badArg');
    });
  });
  describe('Infrastructure', () => {
    it('using Infrastructure constructor overwrites the default Deployment', () => {
      const namespace = 'testing-namespace';
      const machine = new b.Machine({
        provider: 'Amazon',
        region: 'us-west-2',
      });
      const infra = new b.Infrastructure([machine], [machine], { namespace });
      expect(infra.toQuiltRepresentation().namespace).to.equal(namespace);
    });
    it('master and worker machines added correctly', () => {
      const machine = new b.Machine({
        provider: 'Amazon',
        region: 'us-west-2',
      });
      deployment = new b.Infrastructure([machine], [machine, machine]);
      checkMachines([{
        role: 'Master',
        provider: 'Amazon',
        region: 'us-west-2',
      }, {
        // The ID is included here because otherwise the containSubset function
        // used in checkMachines will return true, even if there is only one
        // worker and two masters in the actual output.
        id: '581b4508454ed983d027d699777004feb5c28de3',
        role: 'Worker',
        provider: 'Amazon',
        region: 'us-west-2',
      }, {
        id: '2f822f1272e60e357b679cc520134dbc38d40daa',
        role: 'Worker',
        provider: 'Amazon',
        region: 'us-west-2',
      }]);
    });
    it('errors when no masters are given', () => {
      const machine = new b.Machine({
        provider: 'Amazon',
      });
      expect(() => new b.Infrastructure([], [machine]))
        .to.throw('masters must include 1 or more');
    });
    it('errors when no workers are given', () => {
      const machine = new b.Machine({
        provider: 'Amazon',
      });
      expect(() => new b.Infrastructure([machine], []))
        .to.throw('workers must include 1 or more');
    });
  });
  describe('Query', () => {
    it('namespace', () => {
      deployment = new b.Deployment({ namespace: 'mynamespace' });
      expect(deployment.toQuiltRepresentation().namespace).to.equal(
        'mynamespace');
    });
    it('default namespace', () => {
      expect(deployment.toQuiltRepresentation().namespace).to.equal(
        'default-namespace');
    });
    it('admin ACL', () => {
      deployment = new b.Deployment({ adminACL: ['local'] });
      expect(deployment.toQuiltRepresentation().adminACL).to.eql(
        ['local']);
    });
    it('default admin ACL', () => {
      expect(deployment.toQuiltRepresentation().adminACL).to.eql([]);
    });
  });
  describe('githubKeys()', () => {});
  describe('baseInfrastructure()', () => {
    let fsExistsStub;
    let revertFs;

    beforeEach(() => {
      fsExistsStub = sinon.stub();
      const fsMock = {
        existsSync: fsExistsStub,
      };
      revertFs = b.__set__('fs', fsMock);
    });

    afterEach(() => {
      revertFs();
      fsExistsStub.resetBehavior();
    });

    it('should error if name is not a string', () => {
      const expectedFail = () => {
        b.baseInfrastructure(1);
      };
      expect(expectedFail).to.throw('name must be a string');
    });

    it('should error when the infrastructure doesn\'t exist', () => {
      fsExistsStub.withArgs(b.getInfraPath('foo')).returns(false);
      const expectedFail = () => {
        b.baseInfrastructure('foo');
      };
      expect(expectedFail).to.throw('no infrastructure called foo');
    });

    it('should return the deployment object when the infra exists', () => {
      const expected = 'someDeployment';
      const infraPath = b.getInfraPath('foo');

      const getInfraStub = sinon.stub();
      getInfraStub.withArgs(infraPath).returns(expected);
      const revertGetInfra = b.__set__('getInfraDeployment', getInfraStub);

      fsExistsStub.withArgs(infraPath).returns(true);

      let output;
      const expectedPass = () => {
        output = b.baseInfrastructure('foo');
      };

      expect(expectedPass).to.not.throw();
      expect(output).to.equal(expected);

      revertGetInfra();
    });
  });
});
