package db

import (
	"fmt"
	"net/netip"
)

/*
DBAdapter is the interface for database operations.
Associated Requests are automatically deleted on DelClient().
Deletion of Resources and Clients are managed manually,
with the exception that Resources can't be deleted if still referred to by a Request.

DBAdapter should be safe for concurrent use.

TODO: implement transactions for atomicity
*/
type DBAdapter interface {
	Close() error
	Info() string

	GetResource(url string) (Resource, error)
	PutResource(res Resource) error
	DelResource(url string) error
	FlushExpiredResources() error

	GetClient(addr netip.Addr) (Client, error)
	PutClient(cli Client) error
	DelClient(addr netip.Addr) error
	ListClients() ([]Client, error)
	FlushExpiredClients() error

	GetRequest(addr netip.Addr, url string) (Request, error)
	PutRequest(req Request) error
	DelRequest(addr netip.Addr, url string) error
	ListRequests(addr netip.Addr) ([]Request, error)
	FilterRequests(minSendRatio float64, maturationThreshold int) ([]Request, error) // Order by SendRatio
	FlushExpiredRequests() error

	GetRule(prefix netip.Prefix) (Rule, error)
	PutRule(rule Rule) error
	DelRule(prefix netip.Prefix) error
	ListRules() ([]Rule, error)
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
