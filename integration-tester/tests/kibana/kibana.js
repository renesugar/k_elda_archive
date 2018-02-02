const kelda = require('kelda');
const { Kibana } = require('@kelda/kibana');
const infrastructure = require('../../config/infrastructure.js');

const infra = infrastructure.createTestInfrastructure();

// The Elasticsearch domain was created under Kevin's AWS account according to
// the instructions in the kelda/kibana README.
const elasticsearchURL = 'https://search-kelda-test-zghtwerana5jgig5qvzhpfr7cm.' +
  'us-east-1.es.amazonaws.com';
const elasticsearchPort = 443;
const elasticsearchURLWithPort = `${elasticsearchURL}:${elasticsearchPort}`;

const plugins = [
  'https://github.com/kklin/kbn_dotplot/releases/download/6.0.1/kbn_dotplot.zip',
];

const kibana = new Kibana(elasticsearchURLWithPort, plugins);
kelda.allowTraffic(kibana, kelda.publicInternet, elasticsearchPort);
kibana.deploy(infra);
