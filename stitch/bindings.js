/* eslint require-jsdoc: [1] valid-jsdoc: [1] */
const crypto = require('crypto');
const request = require('sync-request');
const stringify = require('json-stable-stringify');
const _ = require('underscore');
const path = require('path');
const os = require('os');

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

// Both infraDirectory and getInfraPath are also defined in initializer.js.
// This code duplication is ugly, but it significantly simplifies packaging
// the `quilt init` code with the "@quilt/install" module.
const infraDirectory = path.join(os.homedir(), '.quilt', 'infra');

/**
  * Returns the absolute path to the infrastructure with the given name.
  *
  * @param {string} infraName The name of the infrastructure.
  * @return {string} The absolute path to the infrastructure file.
  */
function getInfraPath(infraName) {
  return path.join(infraDirectory, `${infraName}.js`);
}

/**
  * Returns a base infrastructure. The infrastructure can be deployed to simply
  * by calling .deploy() on the returned object.
  * The base infrastructure could be created with `quilt init`.
  *
  * @param {string} name The name of the infrastructure, as passed to
  *   `quilt init`.
  * @return {Deployment} A deployment object representing the infrastructure.
  */
function baseInfrastructure(name = 'default') {
  if (typeof name !== 'string') {
    throw new Error(`name must be a string; was ${stringify(name)}`);
  }

  const infraPath = getInfraPath(name);
  if (!fs.existsSync(infraPath)) {
    throw new Error(`no infrastructure called ${name}. Use 'quilt init' ` +
      `to create a new infrastructure.`);
  }
  const infraGetter = require(infraPath);

  // By passing this module to the infraGetter, the blueprint doesn't have to
  // require Quilt directly and we thus don't have to `npm install` in the
  // infrastructure directory, which would be messy.
  return infraGetter(module.exports);
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

// Returns a globally unique integer ID.
function uniqueID() {
    return uniqueIDCounter++;
}

// setQuiltIDs deterministically sets the id field of objects based on
// their attributes. The _refID field is required to differentiate between
// multiple references to the same object, and multiple instantiations with
// the exact same attributes.
function setQuiltIDs(objs) {
    // The refIDs for each identical instance.
    let refIDs = {};
    objs.forEach(function(obj) {
        let k = obj.hash();
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
        let k = obj.hash();
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
        placements = placements.concat(service.placements);

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

    const quiltDeployment = {
        machines: this.machines,
        labels: services,
        containers: containersNoDups,
        connections: connections,
        placements: placements,

        namespace: this.namespace,
        adminACL: this.adminACL,
        maxPrice: this.maxPrice,
    };
    vet(quiltDeployment);
    return quiltDeployment;
};

// Check if all referenced services in connections and placements are
// really deployed.
function vet(deployment) {
    let labelMap = {[publicInternetLabel]: true};
    deployment.labels.forEach(function(service) {
        labelMap[service.name] = true;
    });

    deployment.connections.forEach((conn) => {
        [conn.from, conn.to].forEach((label) => {
            if (!labelMap[label]) {
                throw new Error(`connection ${stringify(conn)} references ` +
                    `an undeployed service: ${label}`);
            }
        });
    });

    let dockerfiles = {};
    let hostnames = {};
    deployment.containers.forEach((c) => {
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
};

// deploy adds an object, or list of objects, to the deployment.
// Deployable objects must implement the deploy(deployment) interface.
Deployment.prototype.deploy = function(toDeployList) {
    if (!Array.isArray(toDeployList)) {
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
    this.name = uniqueHostname(name);
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

Service.prototype.deploy = function(deployment) {
    deployment.services.push(this);
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
    allowFrom: function(sourceService, portRange) {
        sourceService.allowOutboundPublic(portRange);
    },
};

Service.prototype.allowOutboundPublic = function(range) {
    range = boxRange(range);
    if (range.min != range.max) {
        throw new Error(`public internet can only connect to single ports ` +
            `and not to port ranges`);
    }
    this.outgoingPublic.push(range);
};

Service.prototype.allowFromPublic = function(range) {
    range = boxRange(range);
    if (range.min != range.max) {
        throw new Error(`public internet can only connect to single ports ` +
            `and not to port ranges`);
    }
    this.incomingPublic.push(range);
};

Service.prototype.placeOn = function(machineAttrs) {
    this.placements.push({
        targetLabel: this.name,
        exclusive: false,
        provider: getString('provider', machineAttrs.provider),
        size: getString('size', machineAttrs.size),
        region: getString('region', machineAttrs.region),
        floatingIp: getString('floatingIp', machineAttrs.floatingIp),
    });
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

let hostnameCount = {};
function uniqueHostname(name) {
    if (!(name in hostnameCount)) {
        hostnameCount[name] = 1;
        return name;
    }
    hostnameCount[name]++;
    return uniqueHostname(name + hostnameCount[name]);
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
 * Returns an empty object if `arg` is not defined, and otherwise
 * ensures that `arg` is an object with string keys and values and then returns
 * it.
 */
function getStringMap(argName, arg) {
    if (arg === undefined) {
        return {};
    }
    if (typeof arg !== 'object') {
        throw new Error(`${argName} must be a string map ` +
            `(was: ${stringify(arg)})`);
    }
    Object.keys(arg).forEach((k) => {
        if (typeof k !== 'string') {
            throw new Error(`${argName} must be a string map (key ` +
                `${stringify(k)} is not a string)`);
        }
        if (typeof arg[k] !== 'string') {
            throw new Error(`${argName} must be a string map (value ` +
                `${stringify(arg[k])} associated with ${k} is not a string)`);
        }
    });
    return arg;
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

Machine.prototype.hash = function() {
    return stringify({
        provider: this.provider,
        role: this.role,
        region: this.region,
        size: this.size,
        floatingIp: this.floatingIp,
        diskSize: this.diskSize,
        cpu: this.cpu,
        ram: this.ram,
        preemptible: this.preemptible,
    });
};

function Image(name, dockerfile) {
    this.name = name;
    this.dockerfile = dockerfile;
}

Image.prototype.clone = function() {
    return new Image(this.name, this.dockerfile);
};

function Container(hostnamePrefix, image, {
    command=[],
    env={},
    filepathToContent={}} = {}) {
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

    this.hostnamePrefix = getString('hostnamePrefix', hostnamePrefix);
    this.hostname = uniqueHostname(this.hostnamePrefix);
    this.command = getStringArray('command', command);
    this.env = getStringMap('env', env);
    this.filepathToContent = getStringMap('filepathToContent',
        filepathToContent);

    // Don't allow callers to modify the arguments by reference.
    this.command = _.clone(this.command);
    this.env = _.clone(this.env);
    this.filepathToContent = _.clone(this.filepathToContent);
    this.image = this.image.clone();
}

// Create a new Container with the same attributes.
Container.prototype.clone = function() {
    return new Container(this.hostnamePrefix, this.image, this);
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
    this.hostname = uniqueHostname(h);
};

Container.prototype.getHostname = function() {
    return this.hostname + '.q';
};

Container.prototype.hash = function() {
    return stringify({
        image: this.image,
        command: this.command,
        env: this.env,
        filepathToContent: this.filepathToContent,
        hostname: this.hostname,
    });
};

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
    hostnameCount = {};
}

module.exports = {
    Container,
    Deployment,
    Image,
    Machine,
    Port,
    PortRange,
    Range,
    Service,
    createDeployment,
    getDeployment,
    githubKeys,
    publicInternet,
    resetGlobals,
    baseInfrastructure,
};
