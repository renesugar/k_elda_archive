const chai = require('chai');
const chaiSubset = require('chai-subset');
const {
    Container,
    Image,
    Machine,
    Port,
    PortRange,
    Range,
    Service,
    createDeployment,
    publicInternet,
    resetGlobals,
} = require('./bindings.js');

chai.use(chaiSubset);
const {expect} = chai;

describe('Bindings', function() {
    let deployment;
    beforeEach(function() {
        resetGlobals();
        deployment = createDeployment();
    });

    const checkMachines = function(expected) {
        const {machines} = deployment.toQuiltRepresentation();
        expect(machines).to.have.lengthOf(expected.length)
            .and.containSubset(expected);
    };

    const checkContainers = function(expected) {
        const {containers} = deployment.toQuiltRepresentation();
        expect(containers).to.have.lengthOf(expected.length)
            .and.containSubset(expected);
    };

    const checkPlacements = function(expected) {
        const {placements} = deployment.toQuiltRepresentation();
        expect(placements).to.have.lengthOf(expected.length)
            .and.containSubset(expected);
    };

    const checkLabels = function(expected) {
        const {labels} = deployment.toQuiltRepresentation();
        expect(labels).to.have.lengthOf(expected.length)
            .and.containSubset(expected);
    };

    const checkConnections = function(expected) {
        const {connections} = deployment.toQuiltRepresentation();
        expect(connections).to.have.lengthOf(expected.length)
            .and.containSubset(expected);
    };

    describe('Machine', function() {
        it('basic', function() {
            deployment.deploy([new Machine({
                role: 'Worker',
                provider: 'Amazon',
                region: 'us-west-2',
                size: 'm4.large',
                cpu: new Range(2, 4),
                ram: new Range(4, 8),
                diskSize: 32,
                sshKeys: ['key1', 'key2'],
            })]);
            checkMachines([{
                id: 'ae657514e0aa41ed95d9e27c3f3c9b2ff23bd05e',
                role: 'Worker',
                provider: 'Amazon',
                region: 'us-west-2',
                size: 'm4.large',
                cpu: new Range(2, 4),
                ram: new Range(4, 8),
                diskSize: 32,
                sshKeys: ['key1', 'key2'],
            }]);
        });
        it('hash independent of SSH keys', function() {
            deployment.deploy([new Machine({
                role: 'Worker',
                provider: 'Amazon',
                region: 'us-west-2',
                size: 'm4.large',
                cpu: new Range(2, 4),
                ram: new Range(4, 8),
                diskSize: 32,
                sshKeys: ['key3'],
            })]);
            checkMachines([{
                id: 'ae657514e0aa41ed95d9e27c3f3c9b2ff23bd05e',
                role: 'Worker',
                provider: 'Amazon',
                region: 'us-west-2',
                size: 'm4.large',
                cpu: new Range(2, 4),
                ram: new Range(4, 8),
                diskSize: 32,
                sshKeys: ['key3'],
            }]);
        });
        it('replicate', function() {
            const baseMachine = new Machine({provider: 'Amazon'});
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
        it('replicate independent', function() {
            const baseMachine = new Machine({provider: 'Amazon'});
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
        it('set floating IP', function() {
            const baseMachine = new Machine({
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
        it('preemptible attribute', function() {
            deployment.deploy(new Machine({
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

    describe('Container', function() {
        it('basic', function() {
            deployment.deploy(new Service('foo', [
                new Container('image'),
            ]));
            checkContainers([{
                id: '2d9d1ea3ba5f462202d278f6f07a890c1fb2b8d4',
                image: new Image('image'),
                command: [],
                env: {},
                filepathToContent: {},
            }]);
        });
        it('containers are not duplicated', function() {
            let container = new Container('image');
            deployment.deploy(new Service('foo', [container]));
            deployment.deploy(new Service('bar', [container]));
            checkContainers([{
                id: '2d9d1ea3ba5f462202d278f6f07a890c1fb2b8d4',
                image: new Image('image'),
                command: [],
                env: {},
                filepathToContent: {},
            }]);
        });
        it('command', function() {
            deployment.deploy(new Service('foo', [
                new Container('image', ['arg1', 'arg2']),
            ]));
            checkContainers([{
                id: 'c5dfd3d1747e3600b07781e99c4fb05b4f649b96',
                image: new Image('image'),
                command: ['arg1', 'arg2'],
                env: {},
                filepathToContent: {},
            }]);
        });
        it('env', function() {
            const c = new Container('image');
            c.env.foo = 'bar';
            deployment.deploy(new Service('foo', [c]));
            checkContainers([{
                id: '05ac1a3e606854a5fc2f87b9ea891d0e41d3e6e1',
                image: new Image('image'),
                command: [],
                env: {foo: 'bar'},
                filepathToContent: {},
            }]);
        });
        it('command, env, and files', function() {
            deployment.deploy(new Service('foo', [
                new Container('image', ['arg1', 'arg2'])
                    .withEnv({foo: 'bar'})
                    .withFiles({qux: 'quuz'}),
            ]));
            checkContainers([{
                id: '9d8cfb613ef8df786ac42834b29b19ec1df56a43',
                image: new Image('image'),
                command: ['arg1', 'arg2'],
                env: {foo: 'bar'},
                filepathToContent: {qux: 'quuz'},
            }]);
        });
        it('image dockerfile', function() {
            const c = new Container(new Image('name', 'dockerfile'));
            deployment.deploy(new Service('foo', [c]));
            checkContainers([{
                id: 'ba85ca3e0371189fba2be551a598fe2bbf87a534',
                image: new Image('name', 'dockerfile'),
                command: [],
                env: {},
                filepathToContent: {},
            }]);
        });
        it('replicate', function() {
            deployment.deploy(new Service('foo', new Container('image', ['arg'])
                .replicate(2)));
            checkContainers([
                {
                    id: '667fab4c9692fa0d17af369bb90f4dd6191ed446',
                    image: new Image('image'),
                    command: ['arg'],
                    env: {},
                    filepathToContent: {},
                },
                {
                    id: '50b9741213374695336437b366fca04a7b1541dd',
                    image: new Image('image'),
                    command: ['arg'],
                    env: {},
                    filepathToContent: {},
                },
            ]);
        });
        it('replicate independent', function() {
            const repl = new Container('image', ['arg']).replicate(2);
            repl[0].env.foo = 'bar';
            repl[0].command.push('changed');
            deployment.deploy(new Service('baz', repl));
            checkContainers([
                {
                    id: '667fab4c9692fa0d17af369bb90f4dd6191ed446',
                    image: new Image('image'),
                    command: ['arg'],
                    env: {},
                    filepathToContent: {},
                },
                {
                    id: '2615a8954bbc34b19e4d4f9ba37cd771c29499ac',
                    image: new Image('image'),
                    command: ['arg', 'changed'],
                    env: {foo: 'bar'},
                    filepathToContent: {},
                },
            ]);
        });
        it('hostname', function() {
            const c = new Container(new Image('image'));
            c.setHostname('host');
            deployment.deploy(new Service('foo', [c]));
            checkContainers([{
                id: '293fc7ad8a799d3cf2619a3db7124b0459f395cb',
                image: new Image('image'),
                command: [],
                env: {},
                filepathToContent: {},
                hostname: 'host',
            }]);
        });
        it('#getHostname()', function() {
            const c = new Container('image');
            c.setHostname('host');
            expect(c.getHostname()).to.equal('host.q');
        });
        it('duplicate hostname', function() {
            const a = new Container('image');
            a.hostname = 'host';
            const b = new Container('image');
            b.hostname = 'host';
            deployment.deploy(new Service('foo', [a, b]));
            expect(() => deployment.toQuiltRepresentation()).to
                .throw('hostname "host" used for multiple containers');
        });
        it('setHostname generates unique hostnames', function() {
            const a = new Container('image');
            a.setHostname('host');
            const b = new Container('image');
            b.setHostname('host');
            deployment.deploy(new Service('foo', [a, b]));
            checkContainers([{
                id: '293fc7ad8a799d3cf2619a3db7124b0459f395cb',
                image: new Image('image'),
                command: [],
                env: {},
                filepathToContent: {},
                hostname: 'host',
            }, {
                id: '968bcf8c6d235afbc88aec8e1bdddc506714a0b8',
                image: new Image('image'),
                command: [],
                env: {},
                filepathToContent: {},
                hostname: 'host2',
            }]);
        });
        it('Container.hostname and Service.hostname don\'t conflict', () => {
            const container = new Container('image');
            container.setHostname('foo');
            const serv = new Service('foo', []);
            expect(container.getHostname()).to.equal('foo.q');
            expect(serv.hostname()).to.equal('foo2.q');
        });
    });

    describe('Placement', function() {
        let target;
        beforeEach(function() {
            target = new Service('target', []);
            deployment.deploy(target);
        });
        it('MachineRule size, region, provider', function() {
            target.placeOn({
                size: 'm4.large',
                region: 'us-west-2',
                provider: 'Amazon',
            });
            checkPlacements([{
                targetLabel: 'target',
                exclusive: false,
                region: 'us-west-2',
                provider: 'Amazon',
                size: 'm4.large',
            }]);
        });
        it('MachineRule size, provider', function() {
            target.placeOn({
                size: 'm4.large',
                provider: 'Amazon',
            });
            checkPlacements([{
                targetLabel: 'target',
                exclusive: false,
                provider: 'Amazon',
                size: 'm4.large',
            }]);
        });
        it('MachineRule floatingIp', function() {
            target.placeOn({
                floatingIp: 'xxx.xxx.xxx.xxx',
            });
            checkPlacements([{
                targetLabel: 'target',
                exclusive: false,
                floatingIp: 'xxx.xxx.xxx.xxx',
            }]);
        });
    });
    describe('Label', function() {
        it('basic', function() {
            deployment.deploy(
                new Service('web_tier', [new Container('nginx')]));
            checkLabels([{
                name: 'web_tier',
                ids: ['e08ea919185a436516c87d8dc33342b3adbb2f89'],
            }]);
        });
        it('multiple containers', function() {
            deployment.deploy(new Service('web_tier', [
                new Container('nginx'),
                new Container('nginx'),
            ]));
            checkLabels([{
                name: 'web_tier',
                ids: [
                    'e08ea919185a436516c87d8dc33342b3adbb2f89',
                    '13666ee3835edd19e9ccb840a5c62424cbfd7cea',
                ],
            }]);
        });
        it('duplicate services', function() {
            /* Conflicting label names.  We need to generate a couple of dummy
               containers so that the two deployed containers have _refID's
               that are sorted differently lexicographically and numerically. */
            for (let i = 0; i < 2; i += 1) {
                new Container('image');
            }
            deployment.deploy(new Service('foo', [new Container('image')]));
            for (let i = 0; i < 7; i += 1) {
                new Container('image');
            }
            deployment.deploy(new Service('foo', [new Container('image')]));
            checkLabels([
                {
                    name: 'foo',
                    ids: ['2d9d1ea3ba5f462202d278f6f07a890c1fb2b8d4'],
                },
                {
                    name: 'foo2',
                    ids: ['caedda0972c6ae354de95afc066a9f2fbd2c284b'],
                },
            ]);
        });
        it('get service hostname', function() {
            const foo = new Service('foo', []);
            expect(foo.hostname()).to.equal('foo.q');
        });
        it('get service children', function() {
            const foo = new Service('foo', [
                new Container('bar'),
                new Container('baz'),
            ]);
            expect(foo.children()).to.eql(['1.foo.q', '2.foo.q']);
        });
    });
    describe('AllowFrom', function() {
        let foo;
        let bar;
        beforeEach(function() {
            foo = new Service('foo', []);
            bar = new Service('bar', []);
            deployment.deploy([foo, bar]);
        });
        it('autobox port ranges', function() {
            bar.allowFrom(foo, 80);
            checkConnections([{
                from: 'foo',
                to: 'bar',
                minPort: 80,
                maxPort: 80,
            }]);
        });
        it('port', function() {
            bar.allowFrom(foo, new Port(80));
            checkConnections([{
                from: 'foo',
                to: 'bar',
                minPort: 80,
                maxPort: 80,
            }]);
        });
        it('port range', function() {
            bar.allowFrom(foo, new PortRange(80, 85));
            checkConnections([{
                from: 'foo',
                to: 'bar',
                minPort: 80,
                maxPort: 85,
            }]);
        });
        it('connect to invalid port range', function() {
            expect(() => foo.allowFrom(bar, true)).to
                .throw('Input argument must be a number or a Range');
        });
        it('allow connections to publicInternet', function() {
            publicInternet.allowFrom(foo, 80);
            checkConnections([{
                from: 'foo',
                to: 'public',
                minPort: 80,
                maxPort: 80,
            }]);
        });
        it('allow connections from publicInternet', function() {
            foo.allowFrom(publicInternet, 80);
            checkConnections([{
                from: 'public',
                to: 'foo',
                minPort: 80,
                maxPort: 80,
            }]);
        });
        it('connect to publicInternet port range', function() {
            expect(() =>
                publicInternet.allowFrom(foo, new PortRange(80, 81))).to
                    .throw('public internet can only connect to single ports ' +
                        'and not to port ranges');
        });
        it('connect from publicInternet port range', function() {
            expect(() =>
                foo.allowFrom(publicInternet, new PortRange(80, 81))).to
                    .throw('public internet can only connect to single ports ' +
                        'and not to port ranges');
        });
        it('allowFrom non-service', function() {
            expect(() => foo.allowFrom(10, 10)).to
                .throw(`Services can only connect to other services. ` +
                    `Check that you're allowing connections from a service, ` +
                    `and not from a Container or other object.`);
        });
    });
    describe('Vet', function() {
        let foo;
        const deploy = () => deployment.toQuiltRepresentation();
        beforeEach(function() {
            foo = new Service('foo', []);
            deployment.deploy([foo]);
        });
        it('connect to undeployed label', function() {
            foo.allowFrom(new Service('baz', []), 80);
            expect(deploy).to.throw(
                'connection {"from":"baz","maxPort":80,"minPort":80,'+
                '"to":"foo"} references an undeployed service: baz');
        });
        it('duplicate image', function() {
            deployment.deploy(new Service('foo',
                [new Container(new Image('img', 'dk'))]));
            deployment.deploy(new Service('foo',
                [new Container(new Image('img', 'dk'))]));
            expect(deploy).to.not.throw();
        });
        it('duplicate image with different Dockerfiles', function() {
            deployment.deploy(new Service('foo',
                [new Container(new Image('img', 'dk'))]));
            deployment.deploy(new Service('foo',
                [new Container(new Image('img', 'dk2'))]));
            expect(deploy).to.throw('img has differing Dockerfiles');
        });
    });
    describe('Custom Deploy', function() {
        it('basic', function() {
            deployment.deploy({
                deploy(dep) {
                    dep.deploy([
                        new Service('web_tier', [new Container('nginx')]),
                        new Service('web_tier2', [new Container('nginx')]),
                    ]);
                },
            });
            const {labels} = deployment.toQuiltRepresentation();
            expect(labels).to.have.lengthOf(2)
                .and.containSubset([
                    {
                        name: 'web_tier',
                        ids: ['e08ea919185a436516c87d8dc33342b3adbb2f89'],
                    },
                    {
                        name: 'web_tier2',
                        ids: ['13666ee3835edd19e9ccb840a5c62424cbfd7cea'],
                    },
                ]);
        });
        it('missing deploy', function() {
            expect(() => deployment.deploy({})).to.throw(
                'only objects that implement "deploy(deployment)" can be ' +
                'deployed');
        });
    });
    describe('Create Deployment', function() {
        it('no args', function() {
            expect(createDeployment).to.not.throw();
        });
    });
    describe('Query', function() {
        it('namespace', function() {
            deployment = createDeployment({namespace: 'myNamespace'});
            expect(deployment.toQuiltRepresentation().namespace).to.equal(
                'myNamespace');
        });
        it('default namespace', function() {
            expect(deployment.toQuiltRepresentation().namespace).to.equal(
                'default-namespace');
        });
        it('max price', function() {
            deployment = createDeployment({maxPrice: 5});
            expect(deployment.toQuiltRepresentation().maxPrice).to.equal(5);
        });
        it('default max price', function() {
            expect(deployment.toQuiltRepresentation().maxPrice).to.equal(0);
        });
        it('admin ACL', function() {
            deployment = createDeployment({adminACL: ['local']});
            expect(deployment.toQuiltRepresentation().adminACL).to.eql(
                ['local']);
        });
        it('default admin ACL', function() {
            expect(deployment.toQuiltRepresentation().adminACL).to.eql([]);
        });
    });
    describe('githubKeys()', function() {});
});
