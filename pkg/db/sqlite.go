package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

const versionString = "go-sqlite3 %s (https://github.com/ncruces/go-sqlite3)"

/* Table names */

const (
	sqliteTableClients   = "clients"
	sqliteTableResources = "resources"
	sqliteTableRequests  = "requests"
	sqliteTableRules     = "rules"
)

/* Column names */

const (
	sqliteClientsColAddr      = "addr"
	sqliteClientsColTotalSent = "total_sent"
	sqliteClientsColCreatedOn = "created_on"
	sqliteClientsColExpiresOn = "expires_on"

	sqliteResourcesColURL       = "url"
	sqliteResourcesColSize      = "size"
	sqliteResourcesColExpiresOn = "expires_on"

	sqliteRequestsColAddr      = "addr"
	sqliteRequestsColURL       = "url"
	sqliteRequestsColTotalSent = "total_sent"
	sqliteRequestsColSendRatio = "sent_ratio"
	sqliteRequestsColCreatedOn = "created_on"
	sqliteRequestsColExpiresOn = "expires_on"

	sqliteRulesColPrefix    = "prefix"
	sqliteRulesColAddr      = "addr"
	sqliteRulesColURL       = "url"
	sqliteRulesColExpiresOn = "expires_on"
)

/* SqliteAdapter implements DBAdapter */

type SqliteAdapter struct {
	db *sql.DB
}

func (a *SqliteAdapter) Begin() (DBTx, error) {
	tx, err := a.db.Begin()
	if err != nil {
		return nil, err
	}
	return &SqliteTx{
		tx: tx,
	}, nil
}

/* SqliteTx implements DBTx */

type SqliteTx struct {
	tx *sql.Tx
}

func (t *SqliteTx) Commit() error {
	err := t.tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (t *SqliteTx) Rollback() error {
	err := t.tx.Rollback()
	if err != nil {
		return err
	}
	return nil
}

func makeSqliteAdapter(dbFile string) (*SqliteAdapter, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping sqlite db: %w", err)
	}

	s := &SqliteAdapter{
		db: db,
	}

	err = s.initializeSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return s, nil
}

func (a *SqliteAdapter) initializeSchema() error {
	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer tx.Rollback()

	createClientsTableSQL := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (
	%s TEXT NOT NULL,
	%s INTEGER NOT NULL, %s INTEGER NOT NULL, %s INTEGER NOT NULL, 
	CONSTRAINT clients_pk PRIMARY KEY (%s)
);`,
		sqliteTableClients,
		sqliteClientsColAddr,
		sqliteClientsColTotalSent,
		sqliteClientsColCreatedOn,
		sqliteClientsColExpiresOn,
		sqliteClientsColAddr,
	)

	createResourcesTableSQL := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (
	%s TEXT NOT NULL,
	"%s" INTEGER NOT NULL, "%s" INTEGER NOT NULL,
	CONSTRAINT resources_pk PRIMARY KEY (%s)
);`,
		sqliteTableResources,
		sqliteResourcesColURL,
		sqliteResourcesColSize,
		sqliteResourcesColExpiresOn,
		sqliteResourcesColURL,
	)

	createRequestsTableSQL := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (
	%s TEXT NOT NULL,
	%s TEXT NOT NULL,
	%s INTEGER NOT NULL, %s REAL NOT NULL, %s INTEGER NOT NULL, %s INTEGER NOT NULL,
	UNIQUE (%s, %s),
	CONSTRAINT requests_clients_FK FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE CASCADE,
	CONSTRAINT requests_resources_FK FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE RESTRICT
);`,
		sqliteTableRequests,
		sqliteRequestsColAddr,
		sqliteRequestsColURL,
		sqliteRequestsColTotalSent,
		sqliteRequestsColSendRatio,
		sqliteRequestsColCreatedOn,
		sqliteRequestsColExpiresOn,
		sqliteRequestsColAddr,
		sqliteRequestsColURL,
		sqliteRequestsColAddr,
		sqliteTableClients,
		sqliteClientsColAddr,
		sqliteRequestsColURL,
		sqliteTableResources,
		sqliteResourcesColURL,
	)

	createRulesTableSQL := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (
	%s TEXT NOT NULL,
	%s TEXT,
	%s TEXT,
	%s INTEGER NOT NULL,
	CONSTRAINT rules_pk PRIMARY KEY (%s)
);`,
		sqliteTableRules,
		sqliteRulesColPrefix,
		sqliteRulesColAddr,
		sqliteRulesColURL,
		sqliteRulesColExpiresOn,
		sqliteRulesColPrefix,
	)

	statements := []string{createClientsTableSQL, createResourcesTableSQL, createRequestsTableSQL, createRulesTableSQL}

	for _, v := range statements {
		_, err := tx.Exec(v)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

/* DBAdapter defined functions */

func (a *SqliteAdapter) Close() error {
	return a.db.Close()
}

func (a *SqliteAdapter) Info() string {
	var version string
	a.db.QueryRow(`SELECT sqlite_version()`).Scan(&version)
	return fmt.Sprintf(versionString, version)
}

/* Resources */

func (a *SqliteAdapter) GetResource(url string) (Resource, error) {
	var res Resource
	query := fmt.Sprintf("SELECT %s, %s, %s FROM %s WHERE %s = ?",
		sqliteResourcesColURL,
		sqliteResourcesColSize,
		sqliteResourcesColExpiresOn,
		sqliteTableResources,
		sqliteResourcesColURL,
	)
	var expiresOn int64
	err := a.db.QueryRow(query, url).Scan(&res.URL, &res.Size, &expiresOn)
	if err != nil {
		if err == sql.ErrNoRows {
			return Resource{}, nil // Resource not found
		}
		return Resource{}, fmt.Errorf("failed to get resource: %w", err)
	}
	res.ExpiresOn = time.Unix(expiresOn, 0)
	return res, nil
}

func (t *SqliteTx) PutResource(res Resource) error {
	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s) VALUES (?, ?, ?) ON CONFLICT(%s) DO UPDATE SET %s = excluded.%s, %s = excluded.%s, %s = excluded.%s",
		sqliteTableResources,
		sqliteResourcesColURL,
		sqliteResourcesColSize,
		sqliteResourcesColExpiresOn,
		sqliteResourcesColURL,
		sqliteResourcesColSize,
		sqliteResourcesColSize,
		sqliteResourcesColURL,
		sqliteResourcesColURL,
		sqliteResourcesColExpiresOn,
		sqliteResourcesColExpiresOn,
	)
	_, err := t.tx.Exec(query, res.URL, res.Size, res.ExpiresOn.Unix())
	if err != nil {
		return fmt.Errorf("failed to put resource: %w", err)
	}
	return nil
}

func (t *SqliteTx) DelResource(url string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", sqliteTableResources, sqliteResourcesColURL)
	_, err := t.tx.Exec(query, url)
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}
	return nil
}

func (t *SqliteTx) FlushExpiredResources() error {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s < ?;",
		sqliteTableResources,
		sqliteResourcesColExpiresOn,
	)
	currentTime := time.Now().Unix()
	_, err := t.tx.Exec(query, currentTime)
	// Ignore expired resources referred to by requests
	var sqliteErr sqlite3.ErrorCode
	if err == nil || (errors.As(err, &sqliteErr) && sqliteErr == sqlite3.CONSTRAINT) {
		return nil
	} else {
		return fmt.Errorf("failed to flush expired resources: %w", err)
	}
}

/* Clients */

func (a *SqliteAdapter) GetClient(addr netip.Addr) (Client, error) {
	var cli Client
	query := fmt.Sprintf("SELECT %s, %s, %s, %s FROM %s WHERE %s = ?",
		sqliteClientsColAddr,
		sqliteClientsColTotalSent,
		sqliteClientsColCreatedOn,
		sqliteClientsColExpiresOn,
		sqliteTableClients,
		sqliteClientsColAddr,
	)
	var createdOn, expiresOn int64
	var addrStr string
	err := a.db.QueryRow(query, addr.String()).Scan(&addrStr, &cli.TotalSent, &createdOn, &expiresOn)
	if err != nil {
		if err == sql.ErrNoRows {
			return Client{}, nil // Client not found
		}
		return Client{}, fmt.Errorf("failed to get client: %w", err)
	}
	cli.Addr, err = netip.ParseAddr(addrStr)
	if err != nil {
		return Client{}, fmt.Errorf("failed to parse client addr: %w", err)
	}
	cli.CreatedOn = time.Unix(createdOn, 0)
	cli.ExpiresOn = time.Unix(expiresOn, 0)
	return cli, nil
}

func (t *SqliteTx) PutClient(cli Client) error {
	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES (?, ?, ?, ?) ON CONFLICT(%s) DO UPDATE SET %s = excluded.%s, %s = excluded.%s, %s = excluded.%s, %s = excluded.%s",
		sqliteTableClients,
		sqliteClientsColAddr,
		sqliteClientsColTotalSent,
		sqliteClientsColCreatedOn,
		sqliteClientsColExpiresOn,
		sqliteClientsColAddr,
		sqliteClientsColTotalSent,
		sqliteClientsColTotalSent,
		sqliteClientsColCreatedOn,
		sqliteClientsColCreatedOn,
		sqliteClientsColExpiresOn,
		sqliteClientsColExpiresOn,
		sqliteClientsColAddr,
		sqliteClientsColAddr,
	)
	_, err := t.tx.Exec(query, cli.Addr.String(), cli.TotalSent, cli.CreatedOn.Unix(), cli.ExpiresOn.Unix())
	if err != nil {
		return fmt.Errorf("failed to put client: %w", err)
	}
	return nil
}

func (t *SqliteTx) DelClient(addr netip.Addr) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", sqliteTableClients, sqliteClientsColAddr)
	_, err := t.tx.Exec(query, addr.String())
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}
	return nil
}

func (a *SqliteAdapter) ListClients() ([]Client, error) {
	var clients []Client
	query := fmt.Sprintf("SELECT %s, %s, %s, %s FROM %s",
		sqliteClientsColAddr,
		sqliteClientsColTotalSent,
		sqliteClientsColCreatedOn,
		sqliteClientsColExpiresOn,
		sqliteTableClients,
	)
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var client Client
		var createdOn, expiresOn int64
		var addrStr string
		err := rows.Scan(&addrStr, &client.TotalSent, &createdOn, &expiresOn)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		client.Addr, err = netip.ParseAddr(addrStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse client addr: %w", err)
		}
		client.CreatedOn = time.Unix(createdOn, 0)
		client.ExpiresOn = time.Unix(expiresOn, 0)
		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return clients, nil
}

func (t *SqliteTx) FlushExpiredClients() error {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s < ?;",
		sqliteTableClients,
		sqliteClientsColExpiresOn,
	)
	currentTime := time.Now().Unix()
	_, err := t.tx.Exec(query, currentTime)
	if err != nil {
		return fmt.Errorf("failed to flush expired clients: %w", err)
	}
	return nil
}

/* Requests */

func (a *SqliteAdapter) GetRequest(addr netip.Addr, url string) (Request, error) {
	var req Request
	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s FROM %s WHERE %s = ? AND %s = ?",
		sqliteRequestsColAddr,
		sqliteRequestsColURL,
		sqliteRequestsColTotalSent,
		sqliteRequestsColSendRatio,
		sqliteRequestsColCreatedOn,
		sqliteRequestsColExpiresOn,
		sqliteTableRequests,
		sqliteRequestsColAddr,
		sqliteRequestsColURL,
	)
	var createdOn, expiresOn int64
	var addrStr string
	err := a.db.QueryRow(query, addr.String(), url).Scan(&addrStr, &req.URL, &req.TotalSent, &req.SendRatio, &createdOn, &expiresOn)
	if err != nil {
		if err == sql.ErrNoRows {
			return Request{}, nil // Request not found
		}
		return Request{}, fmt.Errorf("failed to get request: %w", err)
	}
	req.Addr, err = netip.ParseAddr(addrStr)
	if err != nil {
		return Request{}, fmt.Errorf("failed to parse request addr: %w", err)
	}
	req.CreatedOn = time.Unix(createdOn, 0)
	req.ExpiresOn = time.Unix(expiresOn, 0)
	return req, nil
}

func (t *SqliteTx) PutRequest(req Request) error {
	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(%s, %s) DO UPDATE SET %s = excluded.%s, %s = excluded.%s, %s = excluded.%s, %s = excluded.%s",
		sqliteTableRequests,
		sqliteRequestsColAddr,
		sqliteRequestsColURL,
		sqliteRequestsColTotalSent,
		sqliteRequestsColSendRatio,
		sqliteRequestsColCreatedOn,
		sqliteRequestsColExpiresOn,
		sqliteRequestsColAddr,
		sqliteRequestsColURL,
		sqliteRequestsColTotalSent,
		sqliteRequestsColTotalSent,
		sqliteRequestsColSendRatio,
		sqliteRequestsColSendRatio,
		sqliteRequestsColCreatedOn,
		sqliteRequestsColCreatedOn,
		sqliteRequestsColExpiresOn,
		sqliteRequestsColExpiresOn,
	)
	_, err := t.tx.Exec(query, req.Addr.String(), req.URL, req.TotalSent, req.SendRatio, req.CreatedOn.Unix(), req.ExpiresOn.Unix())
	if err != nil {
		return fmt.Errorf("failed to put request: %w", err)
	}
	return nil
}

func (t *SqliteTx) DelRequest(addr netip.Addr, url string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ? AND %s = ?", sqliteTableRequests, sqliteRequestsColAddr, sqliteRequestsColURL)
	_, err := t.tx.Exec(query, addr.String(), url)
	if err != nil {
		return fmt.Errorf("failed to delete request: %w", err)
	}

	return nil
}

func (a *SqliteAdapter) ListRequests(addr netip.Addr) ([]Request, error) {
	var requests []Request
	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s FROM %s WHERE %s = ?",
		sqliteRequestsColAddr,
		sqliteRequestsColURL,
		sqliteRequestsColTotalSent,
		sqliteRequestsColSendRatio,
		sqliteRequestsColCreatedOn,
		sqliteRequestsColExpiresOn,
		sqliteTableRequests,
		sqliteRequestsColAddr,
	)
	rows, err := a.db.Query(query, addr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list requests: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var req Request
		var createdOn, expiresOn int64
		var addrStr string
		err := rows.Scan(&addrStr, &req.URL, &req.TotalSent, &req.SendRatio, &createdOn, &expiresOn)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request: %w", err)
		}
		req.Addr, err = netip.ParseAddr(addrStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse request addr: %w", err)
		}
		req.CreatedOn = time.Unix(createdOn, 0)
		req.ExpiresOn = time.Unix(expiresOn, 0)
		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return requests, nil
}

func (a *SqliteAdapter) FilterRequests(minSendRatio float64, createdBefore time.Time) ([]Request, error) {
	var requests []Request
	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s FROM %s WHERE %s >= ? AND %s < ?",
		sqliteRequestsColAddr,
		sqliteRequestsColURL,
		sqliteRequestsColTotalSent,
		sqliteRequestsColSendRatio,
		sqliteRequestsColCreatedOn,
		sqliteRequestsColExpiresOn,
		sqliteTableRequests,
		sqliteRequestsColSendRatio,
		sqliteClientsColCreatedOn,
	)
	rows, err := a.db.Query(query, minSendRatio, createdBefore.Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to list requests: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var req Request
		var createdOn, expiresOn int64
		var addrStr string
		err := rows.Scan(&addrStr, &req.URL, &req.TotalSent, &req.SendRatio, &createdOn, &expiresOn)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request: %w", err)
		}
		req.Addr, err = netip.ParseAddr(addrStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse client addr: %w", err)
		}
		req.CreatedOn = time.Unix(createdOn, 0)
		req.ExpiresOn = time.Unix(expiresOn, 0)
		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return requests, nil
}

func (t *SqliteTx) FlushExpiredRequests() error {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s < ?;",
		sqliteTableRequests,
		sqliteRequestsColExpiresOn,
	)
	currentTime := time.Now().Unix()
	_, err := t.tx.Exec(query, currentTime)
	if err != nil {
		return fmt.Errorf("failed to flush expired requests: %w", err)
	}
	return nil
}

/* Rules */

func (a *SqliteAdapter) GetRule(prefix netip.Prefix) (dto.Rule, error) {
	var rule dto.Rule
	query := fmt.Sprintf("SELECT %s, %s, %s, %s FROM %s WHERE %s = ?",
		sqliteRulesColPrefix,
		sqliteRulesColAddr,
		sqliteRulesColURL,
		sqliteRulesColExpiresOn,
		sqliteTableRules,
		sqliteRulesColPrefix,
	)
	var expiresOn int64
	var prefixStr, addrStr string
	err := a.db.QueryRow(query, prefix.Masked().String()).Scan(&prefixStr, &addrStr, &rule.URL, &expiresOn)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.Rule{}, nil // Rule not found
		}
		return dto.Rule{}, fmt.Errorf("failed to get rule: %w", err)
	}
	rule.ExpiresOn = time.Unix(expiresOn, 0)
	rule.Prefix, err = netip.ParsePrefix(prefixStr)
	if err != nil {
		return dto.Rule{}, fmt.Errorf("failed to parse rule prefix: %w", err)
	}
	rule.Prefix = rule.Prefix.Masked()
	rule.Addr, err = netip.ParseAddr(addrStr)
	if err != nil {
		return dto.Rule{}, fmt.Errorf("failed to parse rule addr: %w", err)
	}
	return rule, nil
}

func (t *SqliteTx) PutRule(rule dto.Rule) error {
	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES (?, ?, ?, ?) ON CONFLICT(%s) DO UPDATE SET %s = excluded.%s, %s = excluded.%s, %s = excluded.%s, %s = excluded.%s",
		sqliteTableRules,
		sqliteRulesColPrefix,
		sqliteRulesColAddr,
		sqliteRulesColURL,
		sqliteRulesColExpiresOn,
		sqliteRulesColPrefix,
		sqliteRulesColAddr,
		sqliteRulesColAddr,
		sqliteRulesColURL,
		sqliteRulesColURL,
		sqliteRulesColExpiresOn,
		sqliteRulesColExpiresOn,
		sqliteRulesColPrefix,
		sqliteRulesColPrefix,
	)
	_, err := t.tx.Exec(query, rule.Prefix.Masked().String(), rule.Addr.String(), rule.URL, rule.ExpiresOn.Unix())
	if err != nil {
		return fmt.Errorf("failed to put rule: %w", err)
	}
	return nil
}

func (t *SqliteTx) DelRule(prefix netip.Prefix) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", sqliteTableRules, sqliteRulesColPrefix)
	_, err := t.tx.Exec(query, prefix.Masked().String())
	if err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}
	return nil
}

func (a *SqliteAdapter) ListRules() ([]dto.Rule, error) {
	var rules []dto.Rule
	query := fmt.Sprintf("SELECT %s, %s, %s, %s FROM %s",
		sqliteRulesColPrefix,
		sqliteRulesColAddr,
		sqliteRulesColURL,
		sqliteRulesColExpiresOn,
		sqliteTableRules,
	)
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list rules: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rule dto.Rule
		var expiresOn int64
		var prefixStr, addrStr string
		err := rows.Scan(&prefixStr, &addrStr, &rule.URL, &expiresOn)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rule: %w", err)
		}
		rule.ExpiresOn = time.Unix(expiresOn, 0)
		rule.Prefix, err = netip.ParsePrefix(prefixStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rule prefix: %w", err)
		}
		rule.Prefix = rule.Prefix.Masked()
		rule.Addr, err = netip.ParseAddr(addrStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rule addr: %w", err)
		}
		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return rules, nil
}

func (t *SqliteTx) FlushExpiredRules() error {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s < ?;",
		sqliteTableRules,
		sqliteRulesColExpiresOn,
	)
	currentTime := time.Now().Unix()
	_, err := t.tx.Exec(query, currentTime)
	if err != nil {
		return fmt.Errorf("failed to flush expired rules: %w", err)
	}
	return nil
}
