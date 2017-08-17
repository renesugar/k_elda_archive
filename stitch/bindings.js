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
    this.containers = new Set();
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
    setQuiltIDs(this.containers);

    let services = [];
    let connections = [];
    let placements = [];
    let containers = [];

    // Convert the services.
    this.services.forEach(function(service) {
        connections = connections.concat(service.getQuiltConnections());
        services.push({
            name: service.name,
            hostnames: service.containers.map((c) => c.hostname),
        });
    });

    this.containers.forEach((c) => {
        connections = connections.concat(c.getQuiltConnections());
        placements = placements.concat(c.getPlacementsWithID());
        containers.push(c.toQuiltRepresentation());
    });

    const quiltDeployment = {
        machines: this.machines,
        labels: services,
        containers: containers,
        connections: connections,
        placements: placements,

        namespace: this.namespace,
        adminACL: this.adminACL,
        maxPrice: this.maxPrice,
    };
    vet(quiltDeployment);
    return quiltDeployment;
};

// Check if all referenced containers in connections and services are
// really deployed.
function vet(deployment) {
    const labelHostnames = deployment.labels.map((l) => l.name);
    const containerHostnames = deployment.containers.map((c) => c.hostname);
    const hostnames = labelHostnames.concat(containerHostnames);

    const hostnameMap = {[publicInternetLabel]: true};
    hostnames.forEach((hostname) => {
        if (hostnameMap[hostname] !== undefined) {
            throw new Error(`hostname "${hostname}" used multiple times`);
        }
        hostnameMap[hostname] = true;
    });

    deployment.connections.forEach((conn) => {
        [conn.from, conn.to].forEach((host) => {
            if (!hostnameMap[host]) {
                throw new Error(`connection ${stringify(conn)} references ` +
                    `an undefined hostname: ${host}`);
            }
        });
    });

    let dockerfiles = {};
    deployment.containers.forEach((c) => {
        let name = c.image.name;
        if (dockerfiles[name] != undefined &&
                dockerfiles[name] != c.image.dockerfile) {
            throw new Error(`${name} has differing Dockerfiles`);
        }
        dockerfiles[name] = c.image.dockerfile;
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

/**
 * @implements {Connectable}
 */
function Service(name, containers) {
    if (typeof name !== 'string') {
        throw new Error(`name must be a string; was ${stringify(name)}`);
    }
    this.name = uniqueHostname(name);
    this.containers = boxContainers(containers);

    this.allowedInboundConnections = [];
}

// Get the Quilt hostname that represents the entire service.
Service.prototype.hostname = function() {
    return this.name + '.q';
};

Service.prototype.deploy = function(deployment) {
    deployment.services.push(this);
};

/**
 * Allow inbound connections to the load balancer. Note that this does not
 * allow direct connections to the containers behind the load balancer.
 *
 * @param {Container|Container[]} srcArg The containers that can open
 * connections to this Service.
 * @param {int|Port|PortRange} portRange The ports on which containers can open
 * connections.
 * @return {void}
 */
Service.prototype.allowFrom = function(srcArg, portRange) {
    let src;
    try {
        src = boxContainers(srcArg);
    } catch (err) {
        throw new Error(`Services can only allow traffic from containers. ` +
            `Check that you're allowing connections from a Container ` +
            `or list of containers and not from a Service or other object.`);
    }

    src.forEach((c) => {
        this.allowedInboundConnections.push(
            new Connection(c, boxRange(portRange)));
    });
};

// publicInternet is an object that looks like another container that can
// allow inbound connections. However, it is actually just syntactic sugar
// to hide the allowOutboundPublic and allowFromPublic functions.
/**
 * @implements {Connectable}
 */
let publicInternet = {
    allowFrom: function(srcArg, portRange) {
        let src;
        try {
            src = boxContainers(srcArg);
        } catch (err) {
            throw new Error(`Only containers can connect to public. ` +
                `Check that you're allowing connections from a Container or ` +
                `list of containers and not from a Service or other object.`);
        }

        src.forEach((c) => {
          c.allowOutboundPublic(portRange);
        });
    },
};

Service.prototype.getQuiltConnections = function() {
    return this.allowedInboundConnections.map((conn) => {
        return {
            from: conn.from.hostname,
            to: this.name,
            minPort: conn.minPort,
            maxPort: conn.maxPort,
        };
    });
};

function boxContainers(x) {
    if (x instanceof Container) {
        return [x];
    }

    assertContainerList(x);
    return x;
}

function assertContainerList(containers) {
    if (!Array.isArray(containers)) {
        throw new Error(`not an array of Containers (was ` +
            `${stringify(containers)})`);
    }
    for (let i = 0; i < containers.length; i++) {
        if (!(containers[i] instanceof Container)) {
            throw new Error(`not an array of Containers; item ` +
                `at index ${i} (${stringify(containers[i])}) is not a ` +
                `Container`);
        }
    }
}

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

/**
 * @implements {Connectable}
 */
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

    // When generating the Quilt deployment JSON object, these placements must
    // be converted using Container.getPlacementsWithID.
    this.placements = [];

    this.allowedInboundConnections = [];
    this.outgoingPublic = [];
    this.incomingPublic = [];
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

Container.prototype.placeOn = function(machineAttrs) {
    this.placements.push({
        exclusive: false,
        provider: getString('provider', machineAttrs.provider),
        size: getString('size', machineAttrs.size),
        region: getString('region', machineAttrs.region),
        floatingIp: getString('floatingIp', machineAttrs.floatingIp),
    });
};

/**
 * Set the targetContainer of the placement rules to be this container. This
 * cannot be done when `placeOn` is called because the container ID is not
 * determined until after all user code has executed.
 */
Container.prototype.getPlacementsWithID = function() {
    return this.placements.map((plcm) => {
        plcm.targetContainer = this.id;
        return plcm;
    });
};

Container.prototype.allowFrom = function(srcArg, portRange) {
    if (srcArg === publicInternet) {
        this.allowFromPublic(portRange);
        return;
    }

    let src;
    try {
        src = boxContainers(srcArg);
    } catch (err) {
        throw new Error(`Containers can only connect to other containers. ` +
            `Check that you're allowing connections from a container or list ` +
            `of containers, and not from a Service or other object.`);
    }

    src.forEach((c) => {
      this.allowedInboundConnections.push(
          new Connection(c, boxRange(portRange)));
    });
};

Container.prototype.allowOutboundPublic = function(range) {
    range = boxRange(range);
    if (range.min != range.max) {
        throw new Error(`public internet can only connect to single ports ` +
            `and not to port ranges`);
    }
    this.outgoingPublic.push(range);
};

Container.prototype.allowFromPublic = function(range) {
    range = boxRange(range);
    if (range.min != range.max) {
        throw new Error(`public internet can only connect to single ports ` +
            `and not to port ranges`);
    }
    this.incomingPublic.push(range);
};

Container.prototype.deploy = function(deployment) {
    deployment.containers.add(this);
};

Container.prototype.getQuiltConnections = function() {
    let connections = [];

    this.allowedInboundConnections.forEach((conn) => {
        connections.push({
            from: conn.from.hostname,
            to: this.hostname,
            minPort: conn.minPort,
            maxPort: conn.maxPort,
        });
    });

    this.outgoingPublic.forEach((rng) => {
        connections.push({
            from: this.hostname,
            to: publicInternetLabel,
            minPort: rng.min,
            maxPort: rng.max,
        });
    });

    this.incomingPublic.forEach((rng) => {
        connections.push({
            from: publicInternetLabel,
            to: this.hostname,
            minPort: rng.min,
            maxPort: rng.max,
        });
    });

    return connections;
};

Container.prototype.toQuiltRepresentation = function() {
    return {
        id: this.id,
        image: this.image,
        command: this.command,
        env: this.env,
        filepathToContent: this.filepathToContent,
        hostname: this.hostname,
    };
};

/**
 * boxConnectable attempts to convert `objects` into an array of objects that
 * define allowFrom.
 * If `objects` is an Array, it asserts that each element is connectable. If
 * it's just a single object, boxConnectable asserts that it is connectable,
 * and if so, returns it as a single-element Array.
 *
 * @param objects
 * @returns {Connectable[]}
 */
function boxConnectable(objects) {
    if (isConnectable(objects)) {
        return [objects];
    }

    if (!Array.isArray(objects)) {
        throw new Error(`not an array of connectable objects (was ` +
            `${stringify(objects)})`);
    }
    objects.forEach((x, i) => {
        if (!isConnectable(x)) {
            throw new Error(
                `item at index ${i} (${stringify(x)}) cannot be connected to`);
        }
    });
    return objects;
}


/**
 * Interface for classes that can allow inbound traffic.
 *
 *  @interface
 */
// Connectable is never used because it's defining an interface for creating
// JsDoc.
// eslint-disable-next-line no-unused-vars
class Connectable {
  /**
   * allowFrom allows traffic from src on port
   *
   * @param {Container} src The container that can initiate connections.
   * @param {int|Port|PortRange} port The ports to allow traffic on.
   * @return {void}
   */
  allowFrom(src, port) {
    throw new Error('not implemented');
  }
}

/**
 * isConnectable returns whether x can allow inbound connections.
 *
 * @param {object} x The object to check
 * @return {boolean} Whether x can be connected to
 */
function isConnectable(x) {
    return typeof(x.allowFrom) === 'function';
}

/**
 * allow is a utility function to allow calling `allowFrom` on groups of
 * containers.
 *
 * @param {Container|publicInternet} src The containers that can
 * initiate a connection.
 * @param {Connectable[]} dst The objects that traffic can be sent to. Examples
 * of connectable objects are Containers, Services, publicInternet, and
 * user-defined objects that implement allowFrom.
 * @param {int|Port|PortRange} ports The ports that traffic is allowed on.
 * @return {void}
 */
function allow(src, dst, port) {
  boxConnectable(dst).forEach((c) => {
    c.allowFrom(src, port);
  });
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
    allow,
    createDeployment,
    getDeployment,
    githubKeys,
    publicInternet,
    resetGlobals,
    baseInfrastructure,
};
