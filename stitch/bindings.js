/* eslint require-jsdoc: [1] valid-jsdoc: [1] */
const crypto = require('crypto');
const request = require('sync-request');
const stringify = require('json-stable-stringify');
const _ = require('underscore');

const githubCache = {};
function githubKeys(user) {
    if (user in githubCache) {
        return githubCache[user];
    }

    const response = request('GET', `https://github.com/${user}.keys`);
    if (response.statusCode >= 300) {
        // Handle any errors.
        throw new Error(
            `HTTP request for ${user}'s github keys failed with error ` +
            `${response.statusCode}`);
    }

    const keys = response.getBody('utf8').trim().split('\n');
    githubCache[user] = keys;

    return keys;
}

// The default deployment object. createDeployment overwrites this.
global._quiltDeployment = new Deployment({});

// The name used to refer to the public internet in the JSON description
// of the network connections (connections to other services are referenced by
// the name of the service, but since the public internet is not a service,
// we need a special label for it).
let publicInternetLabel = 'public';

// Global unique ID counter.
let uniqueIDCounter = 0;

// Overwrite the deployment object with a new one.
function createDeployment(deploymentOpts) {
    global._quiltDeployment = new Deployment(deploymentOpts);
    return global._quiltDeployment;
}

function Deployment(deploymentOpts) {
    deploymentOpts = deploymentOpts || {};

    this.maxPrice = getNumber('maxPrice', deploymentOpts.maxPrice);
    this.namespace = deploymentOpts.namespace || 'default-namespace';
    this.adminACL = getStringArray('adminACL', deploymentOpts.adminACL);

    this.machines = [];
    this.services = [];
}

function omitSSHKey(key, value) {
    if (key == 'sshKeys') {
        return undefined;
    }
    return value;
}

// Returns a globally unique integer ID.
function uniqueID() {
    return uniqueIDCounter++;
}

// key creates a string key for objects that have a _refID, namely Containers
// and Machines.
function key(obj) {
    let keyObj = obj.clone();
    keyObj._refID = '';
    return stringify(keyObj, {replacer: omitSSHKey});
}

// setQuiltIDs deterministically sets the id field of objects based on
// their attributes. The _refID field is required to differentiate between
// multiple references to the same object, and multiple instantiations with
// the exact same attributes.
function setQuiltIDs(objs) {
    // The refIDs for each identical instance.
    let refIDs = {};
    objs.forEach(function(obj) {
        let k = key(obj);
        if (!refIDs[k]) {
            refIDs[k] = [];
        }
        refIDs[k].push(obj._refID);
    });

    // If there are multiple references to the same object, there will be
    // duplicate refIDs.
    Object.keys(refIDs).forEach(function(k) {
        refIDs[k] = _.sortBy(_.uniq(refIDs[k]), _.identity);
    });

    objs.forEach(function(obj) {
        let k = key(obj);
        obj.id = hash(k + refIDs[k].indexOf(obj._refID));
    });
}

function hash(str) {
    const shaSum = crypto.createHash('sha1');
    shaSum.update(str);
    return shaSum.digest('hex');
}

// Convert the deployment to the QRI deployment format.
Deployment.prototype.toQuiltRepresentation = function() {
    this.vet();

    setQuiltIDs(this.machines);

    // List all of the containers in the deployment. This list may contain
    // duplicates; e.g., if the same container is referenced by multiple
    // services.
    let containers = [];
    this.services.forEach(function(serv) {
        serv.containers.forEach(function(c) {
            containers.push(c);
        });
    });
    setQuiltIDs(containers);

    let services = [];
    let connections = [];
    let placements = [];

    // For each service, convert the associated connections and placement rules.
    // Also, aggregate all containers referenced by services.
    this.services.forEach(function(service) {
        connections = connections.concat(service.getQuiltConnections());
        placements = placements.concat(service.getQuiltPlacements());

        // Collect the containers IDs, and add them to the container map.
        let ids = [];
        service.containers.forEach(function(container) {
            ids.push(container.id);
        });

        services.push({
            name: service.name,
            ids: ids,
        });
    });

    // Create a list of unique containers.
    let addedIds = new Set();
    let containersNoDups = [];
    containers.forEach(function(container) {
        if (!addedIds.has(container.id)) {
            addedIds.add(container.id);
            containersNoDups.push(container);
         }
    });

    return {
        machines: this.machines,
        labels: services,
        containers: containersNoDups,
        connections: connections,
        placements: placements,

        namespace: this.namespace,
        adminACL: this.adminACL,
        maxPrice: this.maxPrice,
    };
};

// Check if all referenced services in connections and placements are
// really deployed.
Deployment.prototype.vet = function() {
    let labelMap = {};
    this.services.forEach(function(service) {
        labelMap[service.name] = true;
    });

    let dockerfiles = {};
    let hostnames = {};
    this.services.forEach(function(service) {
        service.allowedInboundConnections.forEach(function(conn) {
            let from = conn.from.name;
            if (!labelMap[from]) {
                throw new Error(`${service.name} allows connections from ` +
                    `an undeployed service: ${from}`);
            }
        });

        let hasFloatingIp = false;
        service.placements.forEach(function(plcm) {
            if (plcm.floatingIp) {
                hasFloatingIp = true;
            }
        });

        if (hasFloatingIp && service.incomingPublic.length
            && service.containers.length > 1) {
            throw new Error(`${service.name} has a floating IP and ` +
                `multiple containers. This is not yet supported.`);
        }

        service.containers.forEach(function(c) {
            let name = c.image.name;
            if (dockerfiles[name] != undefined &&
                    dockerfiles[name] != c.image.dockerfile) {
                throw new Error(`${name} has differing Dockerfiles`);
            }
            dockerfiles[name] = c.image.dockerfile;

            if (c.hostname !== undefined) {
                if (hostnames[c.hostname]) {
                    throw new Error(`hostname "${c.hostname}" used for ` +
                        `multiple containers`);
                }
                hostnames[c.hostname] = true;
            }
        });
    });
};

// deploy adds an object, or list of objects, to the deployment.
// Deployable objects must implement the deploy(deployment) interface.
Deployment.prototype.deploy = function(toDeployList) {
    if (toDeployList.constructor !== Array) {
        toDeployList = [toDeployList];
    }

    let that = this;
    toDeployList.forEach(function(toDeploy) {
        if (!toDeploy.deploy) {
            throw new Error(`only objects that implement ` +
                `"deploy(deployment)" can be deployed`);
        }
        toDeploy.deploy(that);
    });
};

function Service(name, containers) {
    if (typeof name !== 'string') {
        throw new Error(`name must be a string; was ${stringify(name)}`);
    }
    if (!Array.isArray(containers)) {
        throw new Error(`containers must be an array of Containers (was ` +
            `${stringify(containers)})`);
    }
    for (let i = 0; i < containers.length; i++) {
        if (!(containers[i] instanceof Container)) {
            throw new Error(`containers must be an array of Containers; item ` +
                `at index ${i} (${stringify(containers[i])}) is not a ` +
                `Container`);
        }
    }
    this.name = uniqueLabelName(name);
    this.containers = containers;
    this.placements = [];

    this.allowedInboundConnections = [];
    this.outgoingPublic = [];
    this.incomingPublic = [];
}

// Get the Quilt hostname that represents the entire service.
Service.prototype.hostname = function() {
    return this.name + '.q';
};

// Get a list of Quilt hostnames that address the containers within the service.
Service.prototype.children = function() {
    let i;
    let res = [];
    for (i = 1; i < this.containers.length + 1; i++) {
        res.push(i + '.' + this.name + '.q');
    }
    return res;
};

Service.prototype.deploy = function(deployment) {
    deployment.services.push(this);
};

Service.prototype.connect = function(range, to) {
    console.warn('Warning: connect is deprecated; switch to using ' +
        'allowFrom. If you previously used a.connect(5, b), you should ' +
        'now use b.allowFrom(a, 5).');
    if (!(to === publicInternet || to instanceof Service)) {
        throw new Error(`Services can only connect to other services. ` +
            `Check that you're connecting to a service, and not to a ` +
            `Container or other object.`);
    }
    to.allowFrom(this, range);
};

Service.prototype.allowFrom = function(sourceService, portRange) {
    portRange = boxRange(portRange);
    if (sourceService === publicInternet) {
        return this.allowFromPublic(portRange);
    }
    if (!(sourceService instanceof Service)) {
        throw new Error(`Services can only connect to other services. ` +
            `Check that you're allowing connections from a service, and ` +
            `not from a Container or other object.`);
    }
    this.allowedInboundConnections.push(
        new Connection(sourceService, portRange));
};

// publicInternet is an object that looks like another service that can
// allow inbound connections. However, it is actually just syntactic sugar
// to hide the allowOutboundPublic and allowFromPublic functions.
let publicInternet = {
    connect: function(range, to) {
        console.warn('Warning: connect is deprecated; switch to using ' +
            'allowFrom. Instead of publicInternet.connect(port, service), ' +
            'use service.allowFrom(publicInternet, port).');
        to.allowFromPublic(range);
    },
    allowFrom: function(sourceService, portRange) {
        sourceService.allowOutboundPublic(portRange);
    },
};

// Allow outbound traffic from the service to public internet.
Service.prototype.connectToPublic = function(range) {
    console.warn('Warning: connectToPublic is deprecated; switch to using ' +
        'allowOutboundPublic.');
    this.allowOutboundPublic(range);
};

Service.prototype.allowOutboundPublic = function(range) {
    range = boxRange(range);
    if (range.min != range.max) {
        throw new Error(`public internet can only connect to single ports ` +
            `and not to port ranges`);
    }
    this.outgoingPublic.push(range);
};

// Allow inbound traffic from public internet to the service.
Service.prototype.connectFromPublic = function(range) {
    console.warn('Warning: connectFromPublic is deprecated; switch to ' +
        'allowFromPublic');
    this.allowFromPublic(range);
};

Service.prototype.allowFromPublic = function(range) {
    range = boxRange(range);
    if (range.min != range.max) {
        throw new Error(`public internet can only connect to single ports ` +
            `and not to port ranges`);
    }
    this.incomingPublic.push(range);
};

Service.prototype.place = function(rule) {
    this.placements.push(rule);
};

Service.prototype.getQuiltConnections = function() {
    let connections = [];
    let that = this;

    this.allowedInboundConnections.forEach(function(conn) {
        connections.push({
            from: conn.from.name,
            to: that.name,
            minPort: conn.minPort,
            maxPort: conn.maxPort,
        });
    });

    this.outgoingPublic.forEach(function(rng) {
        connections.push({
            from: that.name,
            to: publicInternetLabel,
            minPort: rng.min,
            maxPort: rng.max,
        });
    });

    this.incomingPublic.forEach(function(rng) {
        connections.push({
            from: publicInternetLabel,
            to: that.name,
            minPort: rng.min,
            maxPort: rng.max,
        });
    });

    return connections;
};

Service.prototype.getQuiltPlacements = function() {
    let placements = [];
    let that = this;
    this.placements.forEach(function(placement) {
        placements.push({
            targetLabel: that.name,
            exclusive: placement.exclusive,

            provider: placement.provider || '',
            size: placement.size || '',
            region: placement.region || '',
            floatingIp: placement.floatingIp || '',
        });
    });
    return placements;
};

let labelNameCount = {};
function uniqueLabelName(name) {
    if (!(name in labelNameCount)) {
        labelNameCount[name] = 0;
    }
    let count = ++labelNameCount[name];
    if (count == 1) {
        return name;
    }
    return name + labelNameCount[name];
}

// Box raw integers into range.
function boxRange(x) {
    if (x === undefined) {
        return new Range(0, 0);
    }
    if (typeof x === 'number') {
        return new Range(x, x);
    }
    if (!(x instanceof Range)) {
        throw new Error('Input argument must be a number or a Range');
    }
    return x;
}

/**
 * Returns 0 if `arg` is not defined, and otherwise ensures that `arg`
 * is a number and then returns it.
 */
function getNumber(argName, arg) {
    if (arg === undefined) {
        return 0;
    }
    if (typeof arg === 'number') {
        return arg;
    }
    throw new Error(`${argName} must be a number (was: ${stringify(arg)})`);
}

/**
 * Returns an empty string if `arg` is not defined, and otherwise
 * ensures that `arg` is a string and then returns it.
 */
function getString(argName, arg) {
    if (arg === undefined) {
        return '';
    }
    if (typeof arg === 'string') {
        return arg;
    }
    throw new Error(`${argName} must be a string (was: ${stringify(arg)})`);
}

/**
 * Returns an empty array if `arg` is not defined, and otherwise
 * ensures that `arg` is an array of strings and then returns it.
 */
function getStringArray(argName, arg) {
    if (arg === undefined) {
        return [];
    }
    if (!Array.isArray(arg)) {
        throw new Error(`${argName} must be an array of strings ` +
            `(was: ${stringify(arg)})`);
    }
    for (let i = 0; i < arg.length; i++) {
        if (typeof arg[i] !== 'string') {
            throw new Error(`${argName} must be an array of strings. ` +
                `Item at index ${i} (${stringify(arg[i])}) is not a ` +
                `string.`);
        }
    }
    return arg;
}

/**
 * Returns false if `arg` is not defined, and otherwise ensures
 * that `arg` is a boolean and then returns it.
 */
function getBoolean(argName, arg) {
    if (arg === undefined) {
        return false;
    }
    if (typeof arg === 'boolean') {
        return arg;
    }
    throw new Error(`${argName} must be a boolean (was: ${stringify(arg)})`);
}

function Machine(optionalArgs) {
    this._refID = uniqueID();

    this.provider = getString('provider', optionalArgs.provider);
    this.role = getString('role', optionalArgs.role);
    this.region = getString('region', optionalArgs.region);
    this.size = getString('size', optionalArgs.size);
    this.floatingIp = getString('floatingIp', optionalArgs.floatingIp);
    this.diskSize = getNumber('diskSize', optionalArgs.diskSize);
    this.sshKeys = getStringArray('sshKeys', optionalArgs.sshKeys);
    this.cpu = boxRange(optionalArgs.cpu);
    this.ram = boxRange(optionalArgs.ram);
    this.preemptible = getBoolean('preemptible', optionalArgs.preemptible);
}

Machine.prototype.deploy = function(deployment) {
    deployment.machines.push(this);
};

// Create a new machine with the same attributes.
Machine.prototype.clone = function() {
    // _.clone only creates a shallow copy, so we must clone sshKeys ourselves.
    let keyClone = _.clone(this.sshKeys);
    let cloned = _.clone(this);
    cloned.sshKeys = keyClone;
    return new Machine(cloned);
};

Machine.prototype.withRole = function(role) {
    let copy = this.clone();
    copy.role = role;
    return copy;
};

Machine.prototype.asWorker = function() {
    return this.withRole('Worker');
};

Machine.prototype.asMaster = function() {
    return this.withRole('Master');
};

// Create n new machines with the same attributes.
Machine.prototype.replicate = function(n) {
    let i;
    let res = [];
    for (i = 0; i < n; i++) {
        res.push(this.clone());
    }
    return res;
};

function Image(name, dockerfile) {
    this.name = name;
    this.dockerfile = dockerfile;
}

Image.prototype.clone = function() {
    return new Image(this.name, this.dockerfile);
};

function Container(image, command) {
    // refID is used to distinguish deployments with multiple references to the
    // same container, and deployments with multiple containers with the exact
    // same attributes.
    this._refID = uniqueID();

    this.image = image;
    if (typeof image === 'string') {
        this.image = new Image(image);
    }
    if (!(this.image instanceof Image)) {
        throw new Error(`image must be an Image or string (was ` +
            `${stringify(image)})`);
    }

    this.command = getStringArray('command', command);
    this.env = {};
    this.filepathToContent = {};
}

// Create a new Container with the same attributes.
Container.prototype.clone = function() {
    let cloned = new Container(this.image.clone(), _.clone(this.command));
    cloned.env = _.clone(this.env);
    cloned.filepathToContent = _.clone(this.filepathToContent);
    return cloned;
};

// Create n new Containers with the same attributes.
Container.prototype.replicate = function(n) {
    let i;
    let res = [];
    for (i = 0; i < n; i++) {
        res.push(this.clone());
    }
    return res;
};

Container.prototype.setEnv = function(key, val) {
    this.env[key] = val;
};

Container.prototype.withEnv = function(env) {
    let cloned = this.clone();
    cloned.env = env;
    return cloned;
};

Container.prototype.withFiles = function(fileMap) {
    let cloned = this.clone();
    cloned.filepathToContent = fileMap;
    return cloned;
};

Container.prototype.setHostname = function(h) {
    this.hostname = h;
};

Container.prototype.getHostname = function() {
    if (this.hostname === undefined) {
        throw new Error('no hostname');
    }
    return this.hostname + '.q';
};

function MachineRule(exclusive, optionalArgs) {
    this.exclusive = exclusive;
    if (optionalArgs.provider) {
        this.provider = optionalArgs.provider;
    }
    if (optionalArgs.size) {
        this.size = optionalArgs.size;
    }
    if (optionalArgs.region) {
        this.region = optionalArgs.region;
    }
    if (optionalArgs.floatingIp) {
      this.floatingIp = optionalArgs.floatingIp;
    }
}

function Connection(from, ports) {
    this.minPort = ports.min;
    this.maxPort = ports.max;
    this.from = from;
}

function Range(min, max) {
    this.min = min;
    this.max = max;
}

function Port(p) {
    return new PortRange(p, p);
}

let PortRange = Range;

function getDeployment() {
    return global._quiltDeployment;
}

// Reset global unique counters. Used only for unit testing.
function resetGlobals() {
    uniqueIDCounter = 0;
    labelNameCount = {};
}

module.exports = {
    Container,
    Deployment,
    Image,
    Machine,
    MachineRule,
    Port,
    PortRange,
    Range,
    Service,
    createDeployment,
    getDeployment,
    githubKeys,
    publicInternet,
    resetGlobals,
};
