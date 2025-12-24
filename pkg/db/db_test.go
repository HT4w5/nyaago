package db

import (
	"net/netip"
	"testing"
	"time"

	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func RunSuite(t *testing.T, adapter DBAdapter) {
	t.Run("ResourceCRUD", func(t *testing.T) { testResourceCRUD(t, adapter) })
	t.Run("ClientCRUD", func(t *testing.T) { testClientCRUD(t, adapter) })
	t.Run("RequestConstraints", func(t *testing.T) { testRequestConstraints(t, adapter) })
	t.Run("RuleCRUD", func(t *testing.T) { testRuleCRUD(t, adapter) })
	t.Run("TransactionRollback", func(t *testing.T) { testTransactionRollback(t, adapter) })
	t.Run("ExpirationFlushing", func(t *testing.T) { testExpirationFlushing(t, adapter) })
}

func testResourceCRUD(t *testing.T, adapter DBAdapter) {
	url := "/file.zip"
	res := Resource{
		URL:       url,
		Size:      1024,
		ExpiresOn: time.Now().Add(time.Hour),
	}

	tx, _ := adapter.Begin()
	require.NoError(t, tx.PutResource(res))
	require.NoError(t, tx.Commit())

	got, err := adapter.GetResource(url)
	assert.NoError(t, err)
	assert.Equal(t, res.Size, got.Size)

	tx, _ = adapter.Begin()
	require.NoError(t, tx.DelResource(url))
	require.NoError(t, tx.Commit())

	got, err = adapter.GetResource(url)
	assert.True(t, len(got.URL) == 0 && err == nil, "Resource should be deleted")
}

func testClientCRUD(t *testing.T, adapter DBAdapter) {
	addr := netip.MustParseAddr("10.0.0.1")
	cli := Client{
		Addr:      addr,
		TotalSent: 5000,
		CreatedOn: time.Now(),
		ExpiresOn: time.Now().Add(time.Hour),
	}

	tx, _ := adapter.Begin()
	require.NoError(t, tx.PutClient(cli))
	require.NoError(t, tx.Commit())

	got, err := adapter.GetClient(addr)
	assert.NoError(t, err)
	assert.Equal(t, cli.TotalSent, got.TotalSent)

	tx, _ = adapter.Begin()
	require.NoError(t, tx.DelClient(addr))
	require.NoError(t, tx.Commit())

	got, err = adapter.GetClient(addr)
	assert.True(t, !got.Addr.IsValid() && err == nil, "Client should be deleted")
}

func testRequestConstraints(t *testing.T, adapter DBAdapter) {
	addr := netip.MustParseAddr("192.168.1.50")
	url := "/resource"

	// Setup: Needs existing Client and Resource
	tx, _ := adapter.Begin()
	_ = tx.PutResource(Resource{URL: url, ExpiresOn: time.Now().Add(time.Hour)})
	_ = tx.PutClient(Client{Addr: addr, ExpiresOn: time.Now().Add(time.Hour)})

	req := Request{
		Addr:      addr,
		URL:       url,
		TotalSent: 100,
		CreatedOn: time.Now(),
	}
	require.NoError(t, tx.PutRequest(req))
	require.NoError(t, tx.Commit())

	// Rule: Resources can't be deleted if still referred to by a Request
	tx, _ = adapter.Begin()
	err := tx.DelResource(url)
	assert.Error(t, err, "Should fail to delete resource because of active request")
	_ = tx.Rollback()

	// Rule: Associated Requests are automatically deleted on DelClient()
	tx, _ = adapter.Begin()
	require.NoError(t, tx.DelClient(addr))
	require.NoError(t, tx.Commit())

	got, err := adapter.GetRequest(addr, url)
	assert.True(t, !got.Addr.IsValid() && err == nil, "Request should have been cascaded deleted with client")
}

func testRuleCRUD(t *testing.T, adapter DBAdapter) {
	prefix := netip.MustParsePrefix("192.168.0.0/24")
	rule := dto.Rule{
		Prefix:    prefix,
		Addr:      netip.MustParseAddr("192.168.0.1"),
		URL:       "/local",
		ExpiresOn: time.Now().Add(time.Hour),
	}

	tx, _ := adapter.Begin()
	require.NoError(t, tx.PutRule(rule))
	require.NoError(t, tx.Commit())

	got, err := adapter.GetRule(prefix)
	assert.NoError(t, err)
	assert.Equal(t, rule.URL, got.URL)

	rules, err := adapter.ListRules()
	assert.NoError(t, err)
	assert.NotEmpty(t, rules)
}

func testTransactionRollback(t *testing.T, adapter DBAdapter) {
	url := "/hidden"

	tx, _ := adapter.Begin()
	_ = tx.PutResource(Resource{URL: url, ExpiresOn: time.Now().Add(time.Hour)})
	require.NoError(t, tx.Rollback())

	got, err := adapter.GetResource(url)
	assert.True(t, len(got.URL) == 0 && err == nil, "Data should not be committed after rollback")
}

func testExpirationFlushing(t *testing.T, adapter DBAdapter) {
	url := "/expired"
	expired := time.Now().Add(-time.Minute)

	tx, _ := adapter.Begin()
	_ = tx.PutResource(Resource{URL: url, ExpiresOn: expired})
	require.NoError(t, tx.FlushExpiredResources())
	require.NoError(t, tx.Commit())

	got, err := adapter.GetResource(url)
	assert.True(t, len(got.URL) == 0 && err == nil, "Expired resource should have been flushed")
}

func TestSqliteImplementation(t *testing.T) {
	db, err := MakeDBAdapter("sqlite", "file::memory:")
	if err != nil {
		t.Skip("Sqlite adapter not implemented or available")
		return
	}
	defer db.Close()

	RunSuite(t, db)
}
