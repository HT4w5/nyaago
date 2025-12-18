package db

import (
	"fmt"
	"net/netip"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) DBAdapter {
	dbFile := ":memory:"
	adapter, err := makeSqliteAdapter(dbFile)
	if err != nil {
		t.Fatalf("Failed to create sqlite adapter: %v", err)
	}
	t.Cleanup(func() {
		adapter.Close()
	})
	return adapter
}

func createTestClient() Client {
	addr := netip.MustParseAddr("192.168.1.1")
	return Client{
		Addr:      addr,
		TotalSent: 1000,
		CreatedOn: time.Now(),
		ExpiresOn: time.Now().Add(24 * time.Hour),
	}
}

func createTestResource() Resource {
	url := "/resource"
	return Resource{
		URL:       url,
		Size:      500,
		ExpiresOn: time.Now().Add(24 * time.Hour),
	}
}

func createTestRequest(client Client, resource Resource) Request {
	return Request{
		Addr:       client.Addr,
		URL:        resource.URL,
		TotalSent:  500,
		SendRatio:  1.1,
		Occurrence: 2,
		CreatedOn:  time.Now(),
		ExpiresOn:  time.Now().Add(24 * time.Hour),
	}
}

func createTestRule(client Client, resource Resource) Rule {
	prefix := netip.MustParsePrefix("192.168.1.1/24")
	return Rule{
		Prefix:    prefix,
		Addr:      client.Addr,
		URL:       resource.URL,
		ExpiresOn: time.Now().Add(24 * time.Hour),
	}
}

func TestPutAndGetClient(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()

	err := adapter.PutClient(client)
	if err != nil {
		t.Errorf("Failed to put client: %v", err)
	}

	gotClient, err := adapter.GetClient(client.Addr)
	if err != nil {
		t.Errorf("Failed to get client: %v", err)
	}

	if !client.Equals(gotClient) {
		t.Errorf("Client mismatch:\n%s", client.Diff(gotClient))
	}
}

func TestPutAndGetResource(t *testing.T) {
	adapter := setupTestDB(t)
	resource := createTestResource()

	err := adapter.PutResource(resource)
	if err != nil {
		t.Errorf("Failed to put resource: %v", err)
	}

	gotResource, err := adapter.GetResource(resource.URL)
	if err != nil {
		t.Errorf("Failed to get resource: %v", err)
	}

	if !resource.Equals(gotResource) {
		t.Errorf("Resource mismatch:\n%s", resource.Diff(gotResource))
	}
}

func TestPutAndGetRequest(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()
	resource := createTestResource()
	req := createTestRequest(client, resource)

	// Need to put client and resource first for foreign key constraints
	err := adapter.PutClient(client)
	if err != nil {
		t.Fatalf("Failed to put client: %v", err)
	}
	err = adapter.PutResource(resource)
	if err != nil {
		t.Fatalf("Failed to put resource: %v", err)
	}

	err = adapter.PutRequest(req)
	if err != nil {
		t.Errorf("Failed to put request: %v", err)
	}

	gotRequest, err := adapter.GetRequest(client.Addr, resource.URL)
	if err != nil {
		t.Errorf("Failed to get request: %v", err)
	}

	if !req.Equals(gotRequest) {
		t.Errorf("Request mismatch:\n%s", req.Diff(gotRequest))
	}
}

func TestListClients(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()

	err := adapter.PutClient(client)
	if err != nil {
		t.Fatalf("Failed to put client: %v", err)
	}

	clients, err := adapter.ListClients()
	if err != nil {
		t.Errorf("Failed to list clients: %v", err)
	}

	if len(clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(clients))
	}

	if !client.Equals(clients[0]) {
		t.Errorf("Request mismatch:\n%s", client.Diff(clients[0]))
	}
}

func TestListRequests(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()
	resource := createTestResource()
	req := createTestRequest(client, resource)

	// Need to put client and resource first for foreign key constraints
	err := adapter.PutClient(client)
	if err != nil {
		t.Fatalf("Failed to put client: %v", err)
	}
	err = adapter.PutResource(resource)
	if err != nil {
		t.Fatalf("Failed to put resource: %v", err)
	}

	err = adapter.PutRequest(req)
	if err != nil {
		t.Fatalf("Failed to put request: %v", err)
	}

	requests, err := adapter.ListRequests(client.Addr)
	if err != nil {
		t.Errorf("Failed to list requests: %v", err)
	}

	if len(requests) != 1 {
		t.Errorf("Expected 1 request, got %d", len(requests))
	}

	if !req.Equals(requests[0]) {
		t.Errorf("Request mismatch:\n%s", req.Diff(requests[0]))
	}
}

func TestPutAndGetRule(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()
	resource := createTestResource()
	rule := createTestRule(client, resource)

	err := adapter.PutClient(client)
	if err != nil {
		t.Fatalf("Failed to put client: %v", err)
	}
	err = adapter.PutResource(resource)
	if err != nil {
		t.Fatalf("Failed to put resource: %v", err)
	}

	err = adapter.PutRule(rule)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	gotRule, err := adapter.GetRule(rule.Prefix)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !rule.Equals(gotRule) {
		t.Errorf("Rule mismatch:\n%s", rule.Diff(gotRule))
	}
}

func TestListRules(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()
	resource := createTestResource()
	rule := createTestRule(client, resource)

	err := adapter.PutClient(client)
	if err != nil {
		t.Fatalf("Failed to put client: %v", err)
	}
	err = adapter.PutResource(resource)
	if err != nil {
		t.Fatalf("Failed to put resource: %v", err)
	}

	err = adapter.PutRule(rule)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	rules, err := adapter.ListRules()
	if err != nil {
		t.Errorf("Failed to list rules: %v", err)
	}

	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}

	if !rule.Equals(rules[0]) {
		t.Errorf("Rule mismatch:\n%s", rule.Diff(rules[0]))
	}
}

func TestGetNonexistentClient(t *testing.T) {
	adapter := setupTestDB(t)
	nonexistentAddr := netip.MustParseAddr("192.168.1.2")
	gotClient, err := adapter.GetClient(nonexistentAddr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if gotClient.Addr.IsValid() {
		t.Errorf("Expected invalid address, got %v", gotClient.Addr)
	}
}

func TestGetNonexistentResource(t *testing.T) {
	adapter := setupTestDB(t)
	nonexistentURL := "/nonexistent"
	gotResource, err := adapter.GetResource(nonexistentURL)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if gotResource.URL != "" {
		t.Errorf("Expected empty URL, got %v", gotResource.URL)
	}
}

func TestGetNonexistentRequest(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()
	nonexistentURL := "/nonexistent"
	gotRequest, err := adapter.GetRequest(client.Addr, nonexistentURL)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if gotRequest.Addr.IsValid() {
		t.Errorf("Expected invalid address, got %v", gotRequest.Addr)
	}
}

func TestDeleteClient(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()

	err := adapter.PutClient(client)
	if err != nil {
		t.Fatalf("Failed to put client: %v", err)
	}

	err = adapter.DelClient(client.Addr)
	if err != nil {
		t.Errorf("Failed to delete client: %v", err)
	}

	gotClient, err := adapter.GetClient(client.Addr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if gotClient.Addr.IsValid() {
		t.Errorf("Expected invalid address, got %v", gotClient.Addr)
	}
}

func TestDeleteResource(t *testing.T) {
	adapter := setupTestDB(t)
	resource := createTestResource()

	err := adapter.PutResource(resource)
	if err != nil {
		t.Fatalf("Failed to put resource: %v", err)
	}

	err = adapter.DelResource(resource.URL)
	if err != nil {
		t.Errorf("Failed to delete resource: %v", err)
	}

	gotResource, err := adapter.GetResource(resource.URL)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if gotResource.URL != "" {
		t.Errorf("Expected empty URL, got %v", gotResource.URL)
	}
}

func TestDeleteRequest(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()
	resource := createTestResource()
	req := createTestRequest(client, resource)

	// Need to put client and resource first for foreign key constraints
	err := adapter.PutClient(client)
	if err != nil {
		t.Fatalf("Failed to put client: %v", err)
	}
	err = adapter.PutResource(resource)
	if err != nil {
		t.Fatalf("Failed to put resource: %v", err)
	}

	err = adapter.PutRequest(req)
	if err != nil {
		t.Fatalf("Failed to put request: %v", err)
	}

	err = adapter.DelRequest(client.Addr, resource.URL)
	if err != nil {
		t.Errorf("Failed to delete request: %v", err)
	}

	gotRequest, err := adapter.GetRequest(client.Addr, resource.URL)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if gotRequest.Addr.IsValid() {
		t.Errorf("Expected invalid address, got %v", gotRequest.Addr)
	}
}

func TestDeleteRule(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()
	resource := createTestResource()
	rule := createTestRule(client, resource)

	err := adapter.PutRule(rule)
	if err != nil {
		t.Fatalf("Failed to put resource: %v", err)
	}

	err = adapter.DelRule(rule.Prefix)
	if err != nil {
		t.Errorf("Failed to delete resource: %v", err)
	}

	getRule, err := adapter.GetRule(rule.Prefix)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if getRule.URL != "" {
		t.Errorf("Expected empty URL, got %v", getRule.URL)
	}
}

func TestListRequestsForNonexistentClient(t *testing.T) {
	adapter := setupTestDB(t)
	nonexistentAddr := netip.MustParseAddr("192.168.1.3")
	requests, err := adapter.ListRequests(nonexistentAddr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(requests) != 0 {
		t.Errorf("Expected 0 requests, got %d", len(requests))
	}
}

func TestPutRequestWithoutClient(t *testing.T) {
	adapter := setupTestDB(t)
	resource := createTestResource()
	newAddr := netip.MustParseAddr("192.168.1.4")
	newRequest := Request{
		Addr:       newAddr,
		URL:        resource.URL,
		TotalSent:  500,
		SendRatio:  1.1,
		Occurrence: 2,
		CreatedOn:  time.Now(),
		ExpiresOn:  time.Now().Add(24 * time.Hour),
	}

	// Put resource first
	err := adapter.PutResource(resource)
	if err != nil {
		t.Fatalf("Failed to put resource: %v", err)
	}

	err = adapter.PutRequest(newRequest)
	if err == nil {
		t.Errorf("Expected error, got none")
	}
}

func TestPuttingConflictingRequestReplacesExisting(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()
	resource := createTestResource()

	// Need to put client and resource first for foreign key constraints
	err := adapter.PutClient(client)
	if err != nil {
		t.Fatalf("Failed to put client: %v", err)
	}
	err = adapter.PutResource(resource)
	if err != nil {
		t.Fatalf("Failed to put resource: %v", err)
	}

	req1 := Request{
		Addr:       client.Addr,
		URL:        resource.URL,
		TotalSent:  114514,
		SendRatio:  1.1,
		Occurrence: 2,
		CreatedOn:  time.Now(),
		ExpiresOn:  time.Now().Add(24 * time.Hour),
	}

	req2 := Request{
		Addr:       client.Addr,
		URL:        resource.URL,
		TotalSent:  1919810,
		SendRatio:  1.1,
		Occurrence: 2,
		CreatedOn:  time.Now(),
		ExpiresOn:  time.Now().Add(24 * time.Hour),
	}

	err = adapter.PutRequest(req1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = adapter.PutRequest(req2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	gotRequest, err := adapter.GetRequest(client.Addr, resource.URL)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !req2.Equals(gotRequest) {
		t.Errorf("Request mismatch:\n%s", req2.Diff(gotRequest))
	}
}

func TestClientDeletionResultsInRequestCascade(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()
	resource := createTestResource()

	// Need to put client and resource first for foreign key constraints
	err := adapter.PutClient(client)
	if err != nil {
		t.Fatalf("Failed to put client: %v", err)
	}
	err = adapter.PutResource(resource)
	if err != nil {
		t.Fatalf("Failed to put resource: %v", err)
	}

	req := Request{
		Addr:       client.Addr,
		URL:        resource.URL,
		TotalSent:  114514,
		SendRatio:  1.1,
		Occurrence: 2,
		CreatedOn:  time.Now(),
		ExpiresOn:  time.Now().Add(24 * time.Hour),
	}

	err = adapter.PutRequest(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = adapter.DelClient(client.Addr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	gotRequest, err := adapter.GetRequest(client.Addr, resource.URL)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if gotRequest.Addr.IsValid() {
		t.Errorf("Expected invalid address, got %v", gotRequest.Addr)
	}

	requests, err := adapter.ListRequests(client.Addr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(requests) != 0 {
		t.Errorf("Expected 0 requests, got %d", len(requests))
	}
}

func TestResourceDeletionFailsWithExistingRequest(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()
	resource := createTestResource()

	// Need to put client and resource first for foreign key constraints
	err := adapter.PutClient(client)
	if err != nil {
		t.Fatalf("Failed to put client: %v", err)
	}
	err = adapter.PutResource(resource)
	if err != nil {
		t.Fatalf("Failed to put resource: %v", err)
	}

	req := Request{
		Addr:       client.Addr,
		URL:        resource.URL,
		TotalSent:  114514,
		SendRatio:  1.1,
		Occurrence: 2,
		CreatedOn:  time.Now(),
		ExpiresOn:  time.Now().Add(24 * time.Hour),
	}

	err = adapter.PutRequest(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = adapter.DelResource(resource.URL)
	if err == nil {
		t.Errorf("Expected error, got none")
	}
}

func TestResourceDeletionFailsWithExistingRequestDuplicate(t *testing.T) {
	adapter := setupTestDB(t)
	client := createTestClient()
	resource := createTestResource()

	// Need to put client and resource first for foreign key constraints
	err := adapter.PutClient(client)
	if err != nil {
		t.Fatalf("Failed to put client: %v", err)
	}
	err = adapter.PutResource(resource)
	if err != nil {
		t.Fatalf("Failed to put resource: %v", err)
	}

	req := Request{
		Addr:       client.Addr,
		URL:        resource.URL,
		TotalSent:  114514,
		SendRatio:  1.1,
		Occurrence: 2,
		CreatedOn:  time.Now(),
		ExpiresOn:  time.Now().Add(24 * time.Hour),
	}

	err = adapter.PutRequest(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = adapter.DelResource(resource.URL)
	if err == nil {
		t.Errorf("Expected error, got none")
	}
}

// -- Flush tests --

func createExpiredTestClient() Client {
	addr := netip.MustParseAddr("192.168.1.100")
	return Client{
		Addr:      addr,
		TotalSent: 1000,
		CreatedOn: time.Now().Add(-48 * time.Hour), // Created 2 days ago
		ExpiresOn: time.Now().Add(-24 * time.Hour), // Expired 1 day ago
	}
}

func createExpiredTestResource() Resource {
	url := "/expired-resource"
	return Resource{
		URL:       url,
		Size:      500,
		ExpiresOn: time.Now().Add(-24 * time.Hour), // Expired 1 day ago
	}
}

func createExpiredTestRequest(client Client, resource Resource) Request {
	return Request{
		Addr:       client.Addr,
		URL:        resource.URL,
		TotalSent:  500,
		SendRatio:  1.1,
		Occurrence: 2,
		CreatedOn:  time.Now().Add(-48 * time.Hour), // Created 2 days ago
		ExpiresOn:  time.Now().Add(-24 * time.Hour), // Expired 1 day ago
	}
}

func createExpiredTestRule(client Client, resource Resource) Rule {
	prefix := netip.MustParsePrefix("192.168.100.1/24")
	return Rule{
		Prefix:    prefix,
		Addr:      client.Addr,
		URL:       resource.URL,
		ExpiresOn: time.Now().Add(-24 * time.Hour), // Expired 1 day ago
	}
}

// --- Client Flush Tests ---

func TestFlushExpiredClients_Expired(t *testing.T) {
	adapter := setupTestDB(t)
	expiredClient := createExpiredTestClient()
	if err := adapter.PutClient(expiredClient); err != nil {
		t.Fatalf("Failed to put expired client: %v", err)
	}

	if err := adapter.FlushExpiredClients(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	got, err := adapter.GetClient(expiredClient.Addr)
	if err != nil {
		t.Fatalf("GetClient failed: %v", err)
	}
	if got.Addr.IsValid() {
		t.Errorf("Expired client should have been deleted")
	}
}

func TestFlushExpiredClients_NonExpired(t *testing.T) {
	adapter := setupTestDB(t)
	nonExpiredClient := createTestClient()
	if err := adapter.PutClient(nonExpiredClient); err != nil {
		t.Fatalf("Failed to put non-expired client: %v", err)
	}

	if err := adapter.FlushExpiredClients(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	got, err := adapter.GetClient(nonExpiredClient.Addr)
	if err != nil {
		t.Fatalf("GetClient failed: %v", err)
	}
	if !got.Addr.IsValid() {
		t.Errorf("Non-expired client should still exist")
	}
}

// --- Resource Flush Tests ---

func TestFlushExpiredResources_Expired(t *testing.T) {
	adapter := setupTestDB(t)
	expiredResource := createExpiredTestResource()
	if err := adapter.PutResource(expiredResource); err != nil {
		t.Fatalf("Failed to put expired resource: %v", err)
	}

	if err := adapter.FlushExpiredResources(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	got, err := adapter.GetResource(expiredResource.URL)
	if err != nil {
		t.Fatalf("GetResource failed: %v", err)
	}
	if got.URL != "" {
		t.Errorf("Expired resource should have been deleted")
	}
}

func TestFlushExpiredResources_NonExpired(t *testing.T) {
	adapter := setupTestDB(t)
	nonExpiredResource := createTestResource()
	if err := adapter.PutResource(nonExpiredResource); err != nil {
		t.Fatalf("Failed to put non-expired resource: %v", err)
	}

	if err := adapter.FlushExpiredResources(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	got, err := adapter.GetResource(nonExpiredResource.URL)
	if err != nil {
		t.Fatalf("GetResource failed: %v", err)
	}
	if got.URL == "" {
		t.Errorf("Non-expired resource should still exist")
	}
}

// --- Request Flush Tests ---

func TestFlushExpiredRequests_Expired(t *testing.T) {
	adapter := setupTestDB(t)
	client, resource := createTestClient(), createTestResource()
	_ = adapter.PutClient(client)
	_ = adapter.PutResource(resource)

	expiredRequest := createExpiredTestRequest(client, resource)
	if err := adapter.PutRequest(expiredRequest); err != nil {
		t.Fatalf("Failed to put expired request: %v", err)
	}

	if err := adapter.FlushExpiredRequests(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	got, err := adapter.GetRequest(client.Addr, expiredRequest.URL)
	if err != nil {
		t.Fatalf("GetRequest failed: %v", err)
	}
	if got.Addr.IsValid() {
		t.Errorf("Expired request should have been deleted")
	}
}

func TestFlushExpiredRequests_NonExpired(t *testing.T) {
	adapter := setupTestDB(t)
	client, resource := createTestClient(), createTestResource()
	_ = adapter.PutClient(client)
	_ = adapter.PutResource(resource)

	nonExpiredRequest := createTestRequest(client, resource)
	if err := adapter.PutRequest(nonExpiredRequest); err != nil {
		t.Fatalf("Failed to put non-expired request: %v", err)
	}

	if err := adapter.FlushExpiredRequests(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	got, err := adapter.GetRequest(client.Addr, nonExpiredRequest.URL)
	if err != nil {
		t.Fatalf("GetRequest failed: %v", err)
	}
	if !got.Addr.IsValid() {
		t.Errorf("Non-expired request should still exist")
	}
}

// --- Rule Flush Tests ---

func TestFlushExpiredRules_Expired(t *testing.T) {
	adapter := setupTestDB(t)
	client, resource := createTestClient(), createTestResource()
	_ = adapter.PutClient(client)
	_ = adapter.PutResource(resource)

	expiredRule := createExpiredTestRule(client, resource)
	if err := adapter.PutRule(expiredRule); err != nil {
		t.Fatalf("Failed to put expired rule: %v", err)
	}

	if err := adapter.FlushExpiredRules(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	got, err := adapter.GetRule(expiredRule.Prefix)
	if err != nil {
		t.Fatalf("GetRule failed: %v", err)
	}
	if got.Prefix.IsValid() {
		t.Errorf("Expired rule should have been deleted")
	}
}

func TestFlushExpiredRules_NonExpired(t *testing.T) {
	adapter := setupTestDB(t)
	client, resource := createTestClient(), createTestResource()
	_ = adapter.PutClient(client)
	_ = adapter.PutResource(resource)

	nonExpiredRule := createTestRule(client, resource)
	if err := adapter.PutRule(nonExpiredRule); err != nil {
		t.Fatalf("Failed to put non-expired rule: %v", err)
	}

	if err := adapter.FlushExpiredRules(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	got, err := adapter.GetRule(nonExpiredRule.Prefix)
	if err != nil {
		t.Fatalf("GetRule failed: %v", err)
	}
	if !got.Prefix.IsValid() {
		t.Errorf("Non-expired rule should still exist")
	}
}

func TestFlushExpiredResources_ReferredToByRequest(t *testing.T) {
	adapter := setupTestDB(t)
	expiredResource := createExpiredTestResource()
	if err := adapter.PutResource(expiredResource); err != nil {
		t.Fatalf("Failed to put expired resource: %v", err)
	}

	// Create request referring to resource
	client := createTestClient()
	request := createTestRequest(client, expiredResource)
	if err := adapter.PutClient(client); err != nil {
		t.Fatalf("Failed to put client: %v", err)
	}
	if err := adapter.PutRequest(request); err != nil {
		t.Fatalf("Failed to put request: %v", err)
	}

	if err := adapter.FlushExpiredResources(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	got, err := adapter.GetResource(expiredResource.URL)
	if err != nil {
		t.Fatalf("GetResource failed: %v", err)
	}
	if got.URL == "" {
		t.Errorf("Referred-to resource should NOT have been deleted")
	}
}

// -- Filter requests test --
func TestFilterRequestsBySendRatio(t *testing.T) {
	adapter := setupTestDB(t)
	addr := netip.MustParseAddr("114.51.4.1")
	client := Client{
		Addr:      addr,
		TotalSent: 160,
		CreatedOn: time.Now(),
		ExpiresOn: time.Now().Add(time.Hour),
	}

	// Values here do not represent actual mechanism
	resources := []Resource{
		{
			URL:       "/low",
			Size:      10,
			ExpiresOn: time.Now().Add(time.Hour),
		},
		{
			URL:       "/mid",
			Size:      10,
			ExpiresOn: time.Now().Add(time.Hour),
		},
		{
			URL:       "/high",
			Size:      10,
			ExpiresOn: time.Now().Add(time.Hour),
		},
	}

	for _, res := range resources {
		err := adapter.PutResource(res)
		if err != nil {
			t.Fatalf("Failed to put resource: %v", err)
		}
	}

	err := adapter.PutClient(client)
	if err != nil {
		t.Fatalf("Failed to put client: %v", err)
	}

	requests := []Request{
		{Addr: addr, URL: "/low", TotalSent: 10, SendRatio: 0.38, Occurrence: 3, CreatedOn: time.Now(), ExpiresOn: time.Now().Add(time.Hour)},
		{Addr: addr, URL: "/high", TotalSent: 100, SendRatio: 5.46, Occurrence: 1, CreatedOn: time.Now(), ExpiresOn: time.Now().Add(time.Hour)},
		{Addr: addr, URL: "/mid", TotalSent: 50, SendRatio: 1.56, Occurrence: 2, CreatedOn: time.Now(), ExpiresOn: time.Now().Add(time.Hour)},
	}

	for _, req := range requests {
		err := adapter.PutRequest(req)
		if err != nil {
			t.Fatalf("Failed to seed request data: %v", err)
		}
	}

	// Fetch > 1.45
	gotRequests, err := adapter.FilterRequests(1.45, 2)
	if err != nil {
		t.Errorf("FilterRequests failed: %v", err)
	}

	if len(gotRequests) != 1 {
		t.Errorf("Expected 1 requests, got %d", len(gotRequests))
	}
}

// -- Testing utils --

func (c *Client) Equals(other Client) bool {
	return c.Addr == other.Addr &&
		c.TotalSent == other.TotalSent &&
		c.CreatedOn.Unix() == other.CreatedOn.Unix() &&
		c.ExpiresOn.Unix() == other.ExpiresOn.Unix()
}

func (c *Client) Diff(other Client) string {
	diff := ""
	if c.Addr != other.Addr {
		diff += fmt.Sprintf("Addr: Expected %s, Got %s; ", c.Addr.String(), other.Addr.String())
	}
	if c.TotalSent != other.TotalSent {
		diff += fmt.Sprintf("TotalSent: Expected %d, Got %d; ", c.TotalSent, other.TotalSent)
	}
	// Use Unix() for time comparison
	if !(c.CreatedOn.Unix() == other.CreatedOn.Unix()) {
		diff += fmt.Sprintf("CreatedOn: Expected %s, Got %s; ", c.CreatedOn.Format(time.RFC3339Nano), other.CreatedOn.Format(time.RFC3339Nano))
	}
	if !(c.ExpiresOn.Unix() == other.ExpiresOn.Unix()) {
		diff += fmt.Sprintf("ExpiresOn: Expected %s, Got %s; ", c.ExpiresOn.Format(time.RFC3339Nano), other.ExpiresOn.Format(time.RFC3339Nano))
	}

	if diff == "" {
		return ""
	}
	// Remove trailing "; " and prefix with struct name
	return "client difference: " + diff[:len(diff)-2]
}

func (r *Resource) Equals(other Resource) bool {
	return r.URL == other.URL &&
		r.Size == other.Size &&
		r.ExpiresOn.Unix() == other.ExpiresOn.Unix()
}

func (r *Resource) Diff(other Resource) string {
	diff := ""
	if r.URL != other.URL {
		diff += fmt.Sprintf("URL: Expected %s, Got %s; ", r.URL, other.URL)
	}
	if r.Size != other.Size {
		diff += fmt.Sprintf("Size: Expected %d, Got %d; ", r.Size, other.Size)
	}
	if !(r.ExpiresOn.Unix() == other.ExpiresOn.Unix()) {
		diff += fmt.Sprintf("ExpiresOn: Expected %s, Got %s; ", r.ExpiresOn.Format(time.RFC3339Nano), other.ExpiresOn.Format(time.RFC3339Nano))
	}

	if diff == "" {
		return ""
	}
	return "resource difference: " + diff[:len(diff)-2]
}

func (r *Request) Equals(other Request) bool {
	return r.Addr == other.Addr &&
		r.URL == other.URL &&
		r.TotalSent == other.TotalSent &&
		r.CreatedOn.Unix() == other.CreatedOn.Unix() &&
		r.ExpiresOn.Unix() == other.ExpiresOn.Unix() &&
		r.SendRatio == other.SendRatio &&
		r.Occurrence == other.Occurrence
}

func (r *Request) Diff(other Request) string {
	diff := ""
	if r.Addr != other.Addr {
		diff += fmt.Sprintf("Addr: Expected %s, Got %s; ", r.Addr.String(), other.Addr.String())
	}
	if r.URL != other.URL {
		diff += fmt.Sprintf("URL: Expected %s, Got %s; ", r.URL, other.URL)
	}
	if r.TotalSent != other.TotalSent {
		diff += fmt.Sprintf("TotalSent: Expected %d, Got %d; ", r.TotalSent, other.TotalSent)
	}
	if r.SendRatio != other.SendRatio {
		diff += fmt.Sprintf("Addr: Expected %f, Got %f; ", r.SendRatio, other.SendRatio)
	}
	if r.Occurrence != other.Occurrence {
		diff += fmt.Sprintf("Addr: Expected %d, Got %d; ", r.Occurrence, other.Occurrence)
	}
	// Use Unix() for time comparison
	if !(r.CreatedOn.Unix() == other.CreatedOn.Unix()) {
		diff += fmt.Sprintf("CreatedOn: Expected %s, Got %s; ", r.CreatedOn.Format(time.RFC3339Nano), other.CreatedOn.Format(time.RFC3339Nano))
	}
	if !(r.ExpiresOn.Unix() == other.ExpiresOn.Unix()) {
		diff += fmt.Sprintf("ExpiresOn: Expected %s, Got %s; ", r.ExpiresOn.Format(time.RFC3339Nano), other.ExpiresOn.Format(time.RFC3339Nano))
	}

	if diff == "" {
		return ""
	}
	return "request difference: " + diff[:len(diff)-2]
}

func (r *Rule) Equals(other Rule) bool {
	return r.Addr == other.Addr &&
		r.URL == other.URL &&
		// Compare with masked prefix
		r.Prefix.Masked() == other.Prefix.Masked() &&
		r.ExpiresOn.Unix() == other.ExpiresOn.Unix()
}

func (r *Rule) Diff(other Rule) string {
	diff := ""
	if r.Addr != other.Addr {
		diff += fmt.Sprintf("Addr: Expected %s, Got %s; ", r.Addr.String(), other.Addr.String())
	}
	if r.URL != other.URL {
		diff += fmt.Sprintf("URL: Expected %s, Got %s; ", r.URL, other.URL)
	}
	if r.Prefix.Masked() != other.Prefix.Masked() {
		diff += fmt.Sprintf("Prefix: Expected %s, Got %s; ", r.Prefix.Masked().String(), other.Prefix.Masked().String())
	}
	// Use Unix() for time comparison
	if !(r.ExpiresOn.Unix() == other.ExpiresOn.Unix()) {
		diff += fmt.Sprintf("ExpiresOn: Expected %s, Got %s; ", r.ExpiresOn.Format(time.RFC3339Nano), other.ExpiresOn.Format(time.RFC3339Nano))
	}

	if diff == "" {
		return ""
	}
	return "request difference: " + diff[:len(diff)-2]
}
