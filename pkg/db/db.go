package db

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/HT4w5/nyaago/pkg/dto"
)

/*
DBAdapter is the interface for database operations.
DBTx is a transaction for write operations.

Associated Requests are automatically deleted on DelClient().
Deletion of Resources and Clients are managed manually,
with the exception that Resources can't be deleted if still referred to by a Request.

DBAdapter should be safe for concurrent use.
*/
type DBAdapter interface {
	Begin() (DBTx, error)
	Close() error
	Info() string

	GetResource(url string) (Resource, error)
	GetClient(addr netip.Addr) (Client, error)
	GetRequest(addr netip.Addr, url string) (Request, error)
	ListRequests(createdBefore time.Time) ([]Request, error)
	FilterRequests(minSendRatio float64, createdBefore time.Time) ([]Request, error) // Filter by SendRatio

	GetRule(prefix netip.Prefix) (dto.Rule, error)
	ListRules() ([]dto.Rule, error)
}

type DBTx interface {
	Commit() error
	Rollback() error

	PutResource(res Resource) error
	DelResource(url string) error
	FlushExpiredResources() error

	PutClient(cli Client) error
	DelClient(addr netip.Addr) error
	FlushExpiredClients() error

	PutRequest(req Request) error
	DelRequest(addr netip.Addr, url string) error
	FlushExpiredRequests() error

	PutRule(rule dto.Rule) error
	DelRule(prefix netip.Prefix) error
	FlushExpiredRules() error
}

func MakeDBAdapter(dbType string, dbAccess string) (DBAdapter, error) {
	switch dbType {
	case "sqlite":
		return makeSqliteAdapter(dbAccess)
	default:
		return nil, fmt.Errorf("unsupported db type: %s", dbType)
	}
}
