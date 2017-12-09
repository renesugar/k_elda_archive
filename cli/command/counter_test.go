package command

import (
	"bytes"
	"flag"
	"testing"

	"github.com/kelda/kelda/api/client/mocks"
	"github.com/kelda/kelda/api/pb"
	"github.com/kelda/kelda/db"
	"github.com/stretchr/testify/assert"
)

func TestCounterFlags(t *testing.T) {
	t.Parallel()

	counters := &Counters{}

	flags := &flag.FlagSet{}
	assert.Nil(t, flags.Usage)

	counters.InstallFlags(flags)

	assert.NotNil(t, flags.Usage)
}

func TestParse(t *testing.T) {
	t.Parallel()

	counters := &Counters{}
	assert.Error(t, counters.Parse(nil), "")

	assert.NoError(t, counters.Parse([]string{"host"}))
	assert.Equal(t, []string{"host"}, counters.targets)

	assert.NoError(t, counters.Parse([]string{"host1", "host2"}))
	assert.Equal(t, []string{"host1", "host2"}, counters.targets)
}

func TestRunQueryDaemon(t *testing.T) {
	t.Parallel()

	counters := &Counters{targets: []string{daemonTarget}}
	mock := new(mocks.Client)
	counters.client = mock

	mock.On("QueryCounters").Once().Return(nil, assert.AnError)
	assert.NotZero(t, counters.Run())

	mock.On("QueryCounters").Once().Return(nil, nil)
	assert.Zero(t, counters.Run())
}

func TestRunQueryMinion(t *testing.T) {
	t.Parallel()

	counters := &Counters{targets: []string{"minion"}}
	mock := new(mocks.Client)
	counters.client = mock

	mock.On("QueryContainers").Return(nil, nil)

	// Test we error when there's no matching machine.
	mock.On("QueryMachines").Once().Return(nil, nil)
	assert.NotZero(t, counters.Run())

	// Test we error when QueryMinionCounters fails.
	mock.On("QueryMachines").Return(
		[]db.Machine{{CloudID: "minion", PublicIP: "host"}}, nil)
	mock.On("QueryMinionCounters", "host").Once().Return(nil, assert.AnError)
	assert.NotZero(t, counters.Run())

	// Test success.
	mock.On("QueryMinionCounters", "host").Once().Return(nil, nil)
	assert.Zero(t, counters.Run())
}

// Test that when the all flag is true, the daemon and all minion counters are
// queried.
func TestRunQueryAll(t *testing.T) {
	t.Parallel()

	counters := &Counters{
		all: true,

		// When all is true, the command line arguments are ignored.
		targets: []string{"ignoreme"},
	}
	mock := new(mocks.Client)
	counters.client = mock

	// Needed by the implementation of util.FuzzyLookup.
	mock.On("QueryContainers").Return(nil, nil)

	pubIP := "pubIP"
	mock.On("QueryMachines").Return([]db.Machine{
		{CloudID: "minion", PublicIP: pubIP},
	}, nil)

	// Even if the daemon counter query fails, the minion should still be queried.
	mock.On("QueryCounters").Return(nil, assert.AnError).Once()
	mock.On("QueryMinionCounters", pubIP).Return(nil, nil)

	// The exit code should be 1 because the daemon counter query failed.
	assert.Equal(t, 1, counters.Run())
	mock.AssertExpectations(t)

	// Now that the daemon counter query succeeds, the exit should be zero.
	mock.On("QueryCounters").Return(nil, nil).Once()
	assert.Zero(t, counters.Run())
}

func TestPrintCounters(t *testing.T) {
	t.Parallel()

	counters := []pb.Counter{{
		Pkg:       "PkgA",
		Name:      "NameA",
		Value:     100,
		PrevValue: 56,
	}, {
		Pkg:       "PkgA",
		Name:      "NameBBBBBBBBBBBBBBB",
		Value:     200,
		PrevValue: 200,
	}, {
		Pkg:       "PkgB",
		Name:      "C",
		Value:     300,
		PrevValue: 0}}

	var b bytes.Buffer
	printCounters(&b, "daemon", counters)
	assert.Equal(t, `daemon
COUNTER                  VALUE  DELTA
                                
PkgA                              
    NameA                100    44
    NameBBBBBBBBBBBBBBB  200    0
PkgB                              
    C                    300    300
`, b.String())
}
