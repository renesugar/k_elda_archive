package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEtcd(t *testing.T) {
	t.Parallel()

	conn := New()

	var id int
	var etcd Etcd
	conn.Txn(EtcdTable).Run(func(view Database) error {
		_, err := view.GetEtcd()
		assert.Error(t, err)

		etcd = view.InsertEtcd()
		id = etcd.ID
		etcd.LeaderIP = "foo"
		view.Commit(etcd)

		etcd, err = view.GetEtcd()
		assert.NoError(t, err)
		return nil
	})

	etcds := conn.SelectFromEtcd(func(i Etcd) bool { return true })

	etcd2 := etcds[0]
	assert.Equal(t, etcd, etcd2)

	assert.Equal(t, "foo", etcd.LeaderIP)
	assert.Equal(t, id, etcd.getID())

	assert.Equal(t, "Etcd-1{EtcdIPs=[], Leader=false, LeaderIP=foo}", etcd.String())

	assert.True(t, etcd.less(Etcd{ID: id + 1}))

	assert.False(t, conn.EtcdLeader())

	assert.Equal(t, etcd, etcd2)
}
