package minion

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kelda/kelda/blueprint"
	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/minion/pb"
)

func TestSetMinionConfig(t *testing.T) {
	t.Parallel()
	s := server{db.New()}

	s.Conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		m := view.InsertMinion()
		m.Self = true
		m.Role = db.Master
		view.Commit(m)
		return nil
	})

	bp := blueprint.Blueprint{
		Containers: []blueprint.Container{
			{Hostname: "hostname"},
		},
	}
	cfg := pb.MinionConfig{
		PrivateIP:      "priv",
		Blueprint:      bp.String(),
		Provider:       "provider",
		Size:           "size",
		Region:         "region",
		EtcdMembers:    []string{"etcd1", "etcd2"},
		AuthorizedKeys: []string{"key1", "key2"},
	}
	expMinion := db.Minion{
		ID:             1,
		Self:           true,
		PrivateIP:      "priv",
		Provider:       "provider",
		Role:           db.Master,
		Size:           "size",
		Region:         "region",
		AuthorizedKeys: "key1\nkey2",
	}
	_, err := s.SetMinionConfig(nil, &cfg)
	assert.NoError(t, err)
	checkMinionEquals(t, s.Conn, expMinion)
	checkEtcdEquals(t, s.Conn, db.Etcd{
		ID:      3,
		EtcdIPs: []string{"etcd1", "etcd2"},
	})
	checkBlueprintEquals(t, s.Conn, db.Blueprint{ID: 2, Blueprint: bp})

	// Update a field.
	bp.Containers[0].Hostname = "changed"
	cfg.Blueprint = bp.String()
	cfg.EtcdMembers = []string{"etcd3"}
	_, err = s.SetMinionConfig(nil, &cfg)
	assert.NoError(t, err)
	checkMinionEquals(t, s.Conn, expMinion)
	checkEtcdEquals(t, s.Conn, db.Etcd{
		ID:      3,
		EtcdIPs: []string{"etcd3"},
	})
	checkBlueprintEquals(t, s.Conn, db.Blueprint{ID: 2, Blueprint: bp})
}

func TestSetMinionMalformedBlueprint(t *testing.T) {
	t.Parallel()

	_, err := server{}.SetMinionConfig(nil, &pb.MinionConfig{Blueprint: "malformed"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to parse blueprint")
}

func checkMinionEquals(t *testing.T, conn db.Conn, exp db.Minion) {
	query := func(conn db.Conn) (interface{}, error) {
		return conn.MinionSelf(), nil
	}
	retryCheckEquals(t, conn, query, exp)
}

func checkEtcdEquals(t *testing.T, conn db.Conn, exp db.Etcd) {
	query := func(conn db.Conn) (row interface{}, err error) {
		conn.Txn(db.EtcdTable).Run(func(view db.Database) error {
			row, err = view.GetEtcd()
			return nil
		})
		return
	}
	retryCheckEquals(t, conn, query, exp)
}

func checkBlueprintEquals(t *testing.T, conn db.Conn, exp db.Blueprint) {
	query := func(conn db.Conn) (row interface{}, err error) {
		conn.Txn(db.BlueprintTable).Run(func(view db.Database) error {
			row, err = view.GetBlueprint()
			return nil
		})
		return
	}
	retryCheckEquals(t, conn, query, exp)
}

type dbQuery func(db.Conn) (interface{}, error)

func retryCheckEquals(t *testing.T, conn db.Conn, query dbQuery, exp interface{}) {
	timeout := time.After(1 * time.Second)
	var actual interface{}
	for {
		var err error
		actual, err = query(conn)
		if err == nil && reflect.DeepEqual(exp, actual) {
			return
		}
		select {
		case <-timeout:
			t.Errorf("Expected database to have %v, but got %v\n",
				exp, actual)
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

}

func TestGetMinionConfig(t *testing.T) {
	t.Parallel()
	s := server{db.New()}

	bp := blueprint.Blueprint{
		Containers: []blueprint.Container{
			{Hostname: "hostname"},
		},
	}
	s.Conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		m := view.InsertMinion()
		m.Self = true
		m.Role = db.Master
		m.PrivateIP = "selfpriv"
		m.Provider = "selfprovider"
		m.Size = "selfsize"
		m.Region = "selfregion"
		m.AuthorizedKeys = "key1\nkey2"
		view.Commit(m)

		bpRow := view.InsertBlueprint()
		bpRow.Blueprint = bp
		view.Commit(bpRow)
		return nil
	})

	// Should only return config for "self".
	s.Conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		m := view.InsertMinion()
		m.Self = false
		m.Role = db.Master
		m.PrivateIP = "priv"
		m.Provider = "provider"
		m.Size = "size"
		m.Region = "region"
		m.AuthorizedKeys = "key1\nkey2"
		view.Commit(m)
		return nil
	})
	cfg, err := s.GetMinionConfig(nil, &pb.Request{})
	assert.NoError(t, err)
	assert.Equal(t, pb.MinionConfig{
		Role:           pb.MinionConfig_MASTER,
		PrivateIP:      "selfpriv",
		Blueprint:      bp.String(),
		Provider:       "selfprovider",
		Size:           "selfsize",
		Region:         "selfregion",
		AuthorizedKeys: []string{"key1", "key2"},
	}, *cfg)

	// Test returning a full config.
	s.Conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		etcd := view.InsertEtcd()
		etcd.EtcdIPs = []string{"etcd1", "etcd2"}
		view.Commit(etcd)
		return nil
	})
	cfg, err = s.GetMinionConfig(nil, &pb.Request{})
	assert.NoError(t, err)
	assert.Equal(t, pb.MinionConfig{
		Role:           pb.MinionConfig_MASTER,
		PrivateIP:      "selfpriv",
		Blueprint:      bp.String(),
		Provider:       "selfprovider",
		Size:           "selfsize",
		Region:         "selfregion",
		EtcdMembers:    []string{"etcd1", "etcd2"},
		AuthorizedKeys: []string{"key1", "key2"},
	}, *cfg)
}
