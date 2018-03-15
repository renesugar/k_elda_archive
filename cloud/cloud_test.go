package cloud

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	"github.com/kelda/kelda/blueprint"
	"github.com/kelda/kelda/cloud/acl"
	"github.com/kelda/kelda/db"
)

var FakeAmazon db.ProviderName = "FakeAmazon"
var FakeVagrant db.ProviderName = "FakeVagrant"
var testRegion = "Fake region"

type fakeProvider struct {
	providerName db.ProviderName
	region       string
	namespace    string
	machines     map[string]db.Machine
	idCounter    int
	cloudConfig  string

	bootRequests []db.Machine
	stopRequests []string
	updatedIPs   []db.Machine
	aclRequests  []acl.ACL

	listError error
}

func fakeValidRegions(p db.ProviderName) []string {
	return []string{testRegion}
}

func (p *fakeProvider) clearLogs() {
	p.bootRequests = nil
	p.stopRequests = nil
	p.aclRequests = nil
	p.updatedIPs = nil
}

func (p *fakeProvider) List() ([]db.Machine, error) {
	if p.listError != nil {
		return nil, p.listError
	}

	var machines []db.Machine
	for _, machine := range p.machines {
		machines = append(machines, machine)
	}
	return machines, nil
}

func (p *fakeProvider) Boot(bootSet []db.Machine) ([]string, error) {
	var ids []string
	for _, toBoot := range bootSet {
		// Record the boot request before we mutate it with implementation
		// details of our fakeProvider.
		p.bootRequests = append(p.bootRequests, toBoot)

		p.idCounter++
		idStr := strconv.Itoa(p.idCounter)
		toBoot.CloudID = idStr
		toBoot.PublicIP = idStr
		ids = append(ids, idStr)
		p.machines[idStr] = toBoot
	}

	return ids, nil
}

func (p *fakeProvider) Stop(machines []db.Machine) error {
	for _, machine := range machines {
		delete(p.machines, machine.CloudID)
		p.stopRequests = append(p.stopRequests, machine.CloudID)
	}
	return nil
}

func (p *fakeProvider) SetACLs(acls []acl.ACL) error {
	p.aclRequests = acls
	return nil
}

func (p *fakeProvider) UpdateFloatingIPs(machines []db.Machine) error {
	for _, desired := range machines {
		curr := p.machines[desired.CloudID]
		curr.FloatingIP = desired.FloatingIP
		p.machines[desired.CloudID] = curr
	}
	p.updatedIPs = append(p.updatedIPs, machines...)
	return nil
}

func (p *fakeProvider) Cleanup() error {
	return nil
}

func newTestCloud(providerName db.ProviderName, region, namespace string) *cloud {
	sleep = func(t time.Duration) {}
	mock()
	cld := cloud{
		conn:         db.New(),
		namespace:    namespace,
		region:       region,
		providerName: providerName,
	}
	provider, _ := newProvider(cld.providerName, cld.namespace, cld.region)
	cld.provider = provider
	return &cld
}

func TestPanicBadProvider(t *testing.T) {
	temp := db.AllProviders
	defer func() {
		r := recover()
		assert.NotNil(t, r)
		db.AllProviders = temp
	}()
	db.AllProviders = []db.ProviderName{FakeAmazon}
	conn := db.New()
	cld := cloud{
		conn:         conn,
		namespace:    "test",
		region:       testRegion,
		providerName: FakeAmazon,
	}
	cld.runOnce()
}

func TestCloudRunOnceInitializesProvider(t *testing.T) {
	mock()
	provider := FakeAmazon
	cld := cloud{
		conn:         db.New(),
		namespace:    "test",
		region:       testRegion,
		providerName: provider,
	}
	cld.runOnce()

	assert.NotNil(t, cld.provider)
	assert.Equal(t, provider, cld.provider.(*fakeProvider).providerName)
}

func TestCloudRunOnceProviderInitializationFailure(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	hook := logrusTest.NewGlobal()

	// Mock newProvider (which is called by runOnce) to return an error.
	errorMessage := "test error message"
	newProvider = func(p db.ProviderName, namespace,
		region string) (provider, error) {
		// Return a Provider, so we can make sure that runOnce doesn't
		// use the returned provider when there's an error.
		return &fakeProvider{}, errors.New(errorMessage)
	}

	providerName := FakeAmazon
	conn := db.New()

	// Insert a blueprint in the database that uses the same provider
	// as the test cloud.
	conn.Txn(db.BlueprintTable).Run(func(view db.Database) error {
		bp := view.InsertBlueprint()
		bp.Blueprint.Machines = []blueprint.Machine{{
			Provider: string(providerName),
			Region:   testRegion,
			Size:     "1",
		}}
		view.Commit(bp)
		return nil
	})

	cld := cloud{
		conn:         conn,
		namespace:    "test",
		region:       testRegion,
		providerName: providerName,
	}
	pollUsedByBlueprint := cld.runOnce()
	assert.Nil(t, cld.provider)

	checkForLogMessage := func(expectedMessage string, level logrus.Level) {
		entryFound := false
		for _, entry := range hook.Entries {
			actualMessage, _ := entry.String()
			if strings.Contains(actualMessage, expectedMessage) {
				assert.Equal(t, level, entry.Level)
				entryFound = true
			}
		}
		assert.True(t, entryFound)
	}

	// Make sure that an error is logged (we can't check that there's just one
	// log entry, because the database will log messages, and depending on the
	// timing, those may appear here).
	checkForLogMessage(errorMessage, logrus.ErrorLevel)

	// Check that if the active blueprint uses a different provider, the message
	// is logged at debug instead of error level.
	hook.Reset()
	conn.Txn(db.BlueprintTable).Run(func(view db.Database) error {
		bp, _ := view.GetBlueprint()
		bp.Blueprint.Machines = []blueprint.Machine{{
			Provider: string(FakeVagrant),
			Region:   testRegion,
			Size:     "1",
		}}
		view.Commit(bp)
		return nil
	})
	pollNotUsedByBlueprint := cld.runOnce()
	assert.Nil(t, cld.provider)
	checkForLogMessage(errorMessage, logrus.DebugLevel)

	assert.True(t, pollUsedByBlueprint < pollNotUsedByBlueprint)
}

func TestCloudRunOnce(t *testing.T) {
	type ipRequest struct {
		id string
		ip string
	}

	type assertion struct {
		boot      []db.Machine
		stop      []string
		updateIPs []ipRequest
	}

	checkSync := func(cld *cloud, expected assertion) {
		cld.runOnce()
		providerInst := cld.provider.(*fakeProvider)

		assert.Equal(t, expected.boot, providerInst.bootRequests, "bootRequests")

		assert.Equal(t, expected.stop, providerInst.stopRequests, "stopRequests")

		var updatedIPs []ipRequest
		for _, m := range providerInst.updatedIPs {
			updatedIPs = append(updatedIPs,
				ipRequest{id: m.CloudID, ip: m.FloatingIP})
		}
		assert.Equal(t, expected.updateIPs, updatedIPs, "updateIPs")

		providerInst.clearLogs()
	}

	var jr joinResult
	cloudJoin = func(cld *cloud) (joinResult, error) { return jr, nil }

	jr.boot = []db.Machine{{
		Provider: FakeAmazon,
		Region:   testRegion,
		Size:     "1",
	}}
	jr.terminate = []db.Machine{{
		Provider: FakeAmazon,
		Region:   testRegion,
		CloudID:  "a",
		Size:     "2",
	}}
	jr.updateIPs = []db.Machine{{
		Provider:   FakeAmazon,
		Region:     testRegion,
		Size:       "1",
		CloudID:    "b",
		FloatingIP: "1.2.3.4",
	}}
	jr.isActive = true

	cld := newTestCloud(FakeAmazon, testRegion, "ns")
	checkSync(cld, assertion{
		boot:      jr.boot,
		updateIPs: []ipRequest{{id: "b", ip: "1.2.3.4"}},
		stop:      []string{"a"},
	})
}

func TestACLs(t *testing.T) {
	myIP = func() (string, error) {
		return "5.6.7.8", nil
	}

	clst := newTestCloud(FakeAmazon, testRegion, "ns")
	clst.syncACLs([]acl.ACL{{CidrIP: "local", MinPort: 80, MaxPort: 80}})

	exp := []acl.ACL{
		{
			CidrIP:  "5.6.7.8/32",
			MinPort: 80,
			MaxPort: 80,
		},
	}
	actual := clst.provider.(*fakeProvider).aclRequests
	assert.Equal(t, exp, actual)
}

func TestStartClouds(t *testing.T) {
	stop := make(chan struct{})
	startClouds(db.New(), "ns", stop)

	// Give the clouds time to be created.
	for i := 0; i < 20 && len(instantiatedProviders) < 3; i++ {
		time.Sleep(500 * time.Millisecond)
	}

	var locations []string
	for _, p := range instantiatedProviders {
		loc := fmt.Sprintf("%s-%s-%s", p.providerName, p.region, p.namespace)
		locations = append(locations, loc)
	}
	sort.Strings(locations)

	// Verify that each cloud provider gets instantiated.
	assert.Equal(t, []string{
		"FakeAmazon-Fake region-ns",
		"FakeAmazon-Fake region-ns",
		"FakeVagrant-Fake region-ns"}, locations)
	close(stop)
}

func TestNewProviderFailure(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("provider.New did not panic on invalid provider")
		}
	}()
	newProviderImpl("FakeAmazon", testRegion, "namespace")
}

func TestDesiredMachines(t *testing.T) {
	cld := newTestCloud(FakeAmazon, testRegion, "ns")
	adminKey = "bar"

	res := cld.desiredMachines([]blueprint.Machine{{
		Provider: "Google", // Wrong Provider
		Region:   "zone-1",
	}, {
		Provider: string(FakeAmazon),
		Region:   testRegion,
		Role:     "invalid",
	}, {
		Provider:    string(FakeAmazon),
		Region:      testRegion,
		Size:        "m4.lage",
		Preemptible: true,
		FloatingIP:  "1.2.3.4",
		Role:        db.Worker,
		SSHKeys:     []string{"foo"},
	}})
	assert.Equal(t, []db.Machine{{
		Provider:    FakeAmazon,
		Region:      testRegion,
		Size:        "m4.lage",
		Preemptible: true,
		FloatingIP:  "1.2.3.4",
		Role:        db.Worker,
		DiskSize:    defaultDiskSize,
		SSHKeys:     []string{"foo", "bar"}}}, res)
}

func TestRunOnceMaxPoll(t *testing.T) {
	var jr joinResult
	cloudJoin = func(cld *cloud) (joinResult, error) { return jr, nil }
	cld := newTestCloud(FakeAmazon, testRegion, "ns")

	jr.isActive = false
	pollInactive := cld.runOnce()

	jr.isActive = true
	pollActiveNoChanges := cld.runOnce()

	jr.boot = []db.Machine{{}}
	pollActiveWithChanges := cld.runOnce()

	assert.True(t, pollActiveWithChanges < pollActiveNoChanges,
		"we should poll more often in clouds that are expected to change soon")
	assert.True(t, pollActiveNoChanges < pollInactive,
		"we should poll more often in clouds that have machines running")
}

var instantiatedProviders []fakeProvider

func mock() {
	instantiatedProviders = nil
	var mutex sync.Mutex
	newProvider = func(p db.ProviderName, namespace,
		region string) (provider, error) {

		mutex.Lock()
		defer mutex.Unlock()

		ret := fakeProvider{
			providerName: p,
			region:       region,
			namespace:    namespace,
			machines:     make(map[string]db.Machine),
		}
		ret.clearLogs()

		instantiatedProviders = append(instantiatedProviders, ret)
		return &ret, nil
	}

	ValidRegions = fakeValidRegions
	db.AllProviders = []db.ProviderName{FakeAmazon, FakeVagrant}
}
