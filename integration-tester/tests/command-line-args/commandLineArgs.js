const infrastructure = require('../../config/infrastructure.js');

// The containers are deployed from the Go code in command_line_args_test.go.
infrastructure.createTestInfrastructure();
