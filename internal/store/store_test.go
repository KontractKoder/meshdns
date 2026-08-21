package store

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestOpenTwiceOnSameFile(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "meshdns.db")

	first, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open first = %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close first = %v", err)
	}

	second, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open second = %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatalf("Close second = %v", err)
	}
}

func TestCreateServerGetServerRoundTripAndDuplicateName(t *testing.T) {
	st := newTestStore(t)

	created := testServer("srv-1", "alpha")
	created.Description = "Alpha resolver"
	created.Capabilities = []string{"dns", "doh"}

	if err := st.CreateServer(created, "hash-one"); err != nil {
		t.Fatalf("CreateServer = %v", err)
	}

	got, err := st.GetServer(created.ID)
	if err != nil {
		t.Fatalf("GetServer = %v", err)
	}

	if got.ID != created.ID || got.Name != created.Name || got.Description != created.Description || got.ServerURL != created.ServerURL {
		t.Fatalf("GetServer = %+v, want fields from %+v", got, created)
	}
	if got.Status != "active" {
		t.Fatalf("Status = %q, want active", got.Status)
	}
	if !reflect.DeepEqual(got.Capabilities, []string{"dns", "doh"}) {
		t.Fatalf("Capabilities = %#v, want dns,doh", got.Capabilities)
	}
	if _, err := time.Parse(time.RFC3339, got.CreatedAt); err != nil {
		t.Fatalf("CreatedAt is not RFC3339: %q", got.CreatedAt)
	}
	if _, err := time.Parse(time.RFC3339, got.UpdatedAt); err != nil {
		t.Fatalf("UpdatedAt is not RFC3339: %q", got.UpdatedAt)
	}

	byName, err := st.GetServerByName(created.Name)
	if err != nil {
		t.Fatalf("GetServerByName = %v", err)
	}
	if byName.ID != created.ID {
		t.Fatalf("GetServerByName ID = %q, want %q", byName.ID, created.ID)
	}

	duplicate := testServer("srv-2", created.Name)
	err = st.CreateServer(duplicate, "hash-two")
	if !errors.Is(err, ErrDuplicateName) {
		t.Fatalf("CreateServer duplicate err = %v, want ErrDuplicateName", err)
	}
}

func TestSetCapabilitiesAndListServersCapabilityFilter(t *testing.T) {
	st := newTestStore(t)

	mustCreate(t, st, testServer("srv-1", "alpha"))
	mustCreate(t, st, testServer("srv-2", "bravo"))

	if err := st.SetCapabilities("srv-1", []string{"dns", "doh", "dns"}); err != nil {
		t.Fatalf("SetCapabilities srv-1 = %v", err)
	}
	if err := st.SetCapabilities("srv-2", []string{"metrics"}); err != nil {
		t.Fatalf("SetCapabilities srv-2 = %v", err)
	}

	got, next, err := st.ListServers("", "dns", "active", "", 10)
	if err != nil {
		t.Fatalf("ListServers = %v", err)
	}
	if next != "" {
		t.Fatalf("nextCursor = %q, want empty", next)
	}
	if len(got) != 1 || got[0].ID != "srv-1" {
		t.Fatalf("ListServers capability result = %+v, want srv-1 only", got)
	}
	if !reflect.DeepEqual(got[0].Capabilities, []string{"dns", "doh"}) {
		t.Fatalf("Capabilities = %#v, want sorted unique dns,doh", got[0].Capabilities)
	}
}

func TestListServersQuerySubstringAndPagination(t *testing.T) {
	st := newTestStore(t)

	names := []string{"alpha", "bravo", "charlie", "delta", "echo"}
	for i, name := range names {
		server := testServer(fmt.Sprintf("srv-%d", i+1), name)
		server.Description = fmt.Sprintf("Regional NODE %d", i+1)
		mustCreate(t, st, server)
	}

	var gotNames []string
	cursor := ""
	for {
		page, next, err := st.ListServers("node", "", "active", cursor, 2)
		if err != nil {
			t.Fatalf("ListServers cursor %q = %v", cursor, err)
		}
		for _, server := range page {
			gotNames = append(gotNames, server.Name)
		}
		if next == "" {
			break
		}
		cursor = next
	}

	if !reflect.DeepEqual(gotNames, names) {
		t.Fatalf("paginated names = %#v, want %#v", gotNames, names)
	}
}

func TestRecordProbeAndGetUptime30d(t *testing.T) {
	st := newTestStore(t)
	mustCreate(t, st, testServer("srv-1", "alpha"))

	base := time.Now().UTC().Add(-time.Hour)
	probes := []struct {
		up bool
	}{
		{up: true},
		{up: true},
		{up: true},
		{up: false},
	}

	for i, probe := range probes {
		ts := base.Add(time.Duration(i) * time.Minute).Format(time.RFC3339)
		if err := st.RecordProbe("srv-1", ts, probe.up, 25+i); err != nil {
			t.Fatalf("RecordProbe %d = %v", i, err)
		}
	}

	uptime, err := st.GetUptime30d("srv-1")
	if err != nil {
		t.Fatalf("GetUptime30d = %v", err)
	}
	if uptime != 0.75 {
		t.Fatalf("GetUptime30d = %v, want 0.75", uptime)
	}
}

func TestSetServerStateAndGetUpServersByCapability(t *testing.T) {
	st := newTestStore(t)

	for _, server := range []Server{
		testServer("srv-low", "alpha"),
		testServer("srv-high", "bravo"),
		testServer("srv-down", "charlie"),
		testServer("srv-delisted", "delta"),
	} {
		server.Capabilities = []string{"dns"}
		mustCreate(t, st, server)
	}

	checkedAt := time.Now().UTC().Format(time.RFC3339)
	for _, tc := range []struct {
		id        string
		up        bool
		uptime30d float64
	}{
		{id: "srv-low", up: true, uptime30d: 0.8},
		{id: "srv-high", up: true, uptime30d: 0.95},
		{id: "srv-down", up: false, uptime30d: 1},
		{id: "srv-delisted", up: true, uptime30d: 0.99},
	} {
		if err := st.SetServerState(tc.id, tc.up, checkedAt, tc.uptime30d); err != nil {
			t.Fatalf("SetServerState %s = %v", tc.id, err)
		}
	}
	if err := st.DelistServer("srv-delisted"); err != nil {
		t.Fatalf("DelistServer = %v", err)
	}

	got, err := st.GetUpServersByCapability("dns")
	if err != nil {
		t.Fatalf("GetUpServersByCapability = %v", err)
	}

	gotIDs := serverIDs(got)
	wantIDs := []string{"srv-high", "srv-low"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("GetUpServersByCapability IDs = %#v, want %#v", gotIDs, wantIDs)
	}
}

func TestAppendEventAndCountEventsSince(t *testing.T) {
	st := newTestStore(t)

	base := time.Now().UTC().Add(-2 * time.Hour)
	events := []struct {
		ts        string
		eventType string
	}{
		{ts: base.Format(time.RFC3339), eventType: "server.up"},
		{ts: base.Add(time.Hour).Format(time.RFC3339), eventType: "server.up"},
		{ts: base.Add(2 * time.Hour).Format(time.RFC3339), eventType: "server.down"},
	}
	for _, event := range events {
		if err := st.AppendEvent(event.ts, event.eventType, `{"server_id":"srv-1"}`); err != nil {
			t.Fatalf("AppendEvent = %v", err)
		}
	}

	count, err := st.CountEventsSince("server.up", base.Add(30*time.Minute).Format(time.RFC3339))
	if err != nil {
		t.Fatalf("CountEventsSince = %v", err)
	}
	if count != 1 {
		t.Fatalf("CountEventsSince = %d, want 1", count)
	}
}

func TestExportAllIncludesDelisted(t *testing.T) {
	st := newTestStore(t)

	mustCreate(t, st, testServer("srv-active", "alpha"))
	mustCreate(t, st, testServer("srv-delisted", "bravo"))
	if err := st.DelistServer("srv-delisted"); err != nil {
		t.Fatalf("DelistServer = %v", err)
	}

	got, err := st.ExportAll()
	if err != nil {
		t.Fatalf("ExportAll = %v", err)
	}

	statuses := map[string]string{}
	for _, server := range got {
		statuses[server.ID] = server.Status
	}
	if statuses["srv-active"] != "active" || statuses["srv-delisted"] != "delisted" {
		t.Fatalf("ExportAll statuses = %#v, want active and delisted", statuses)
	}
}

func TestCountServersAndUpServers(t *testing.T) {
	st := newTestStore(t)

	for _, server := range []Server{
		testServer("srv-up", "alpha"),
		testServer("srv-down", "bravo"),
		testServer("srv-delisted", "charlie"),
	} {
		mustCreate(t, st, server)
	}
	if err := st.DelistServer("srv-delisted"); err != nil {
		t.Fatalf("DelistServer = %v", err)
	}

	checkedAt := time.Now().UTC().Format(time.RFC3339)
	if err := st.SetServerState("srv-up", true, checkedAt, 0.9); err != nil {
		t.Fatalf("SetServerState srv-up = %v", err)
	}
	if err := st.SetServerState("srv-down", false, checkedAt, 0.5); err != nil {
		t.Fatalf("SetServerState srv-down = %v", err)
	}
	if err := st.SetServerState("srv-delisted", true, checkedAt, 1); err != nil {
		t.Fatalf("SetServerState srv-delisted = %v", err)
	}

	active, err := st.CountServers("active")
	if err != nil {
		t.Fatalf("CountServers active = %v", err)
	}
	total, err := st.CountServers("")
	if err != nil {
		t.Fatalf("CountServers total = %v", err)
	}
	upActive, err := st.CountUpServers("active")
	if err != nil {
		t.Fatalf("CountUpServers active = %v", err)
	}
	upTotal, err := st.CountUpServers("")
	if err != nil {
		t.Fatalf("CountUpServers total = %v", err)
	}

	if active != 2 || total != 3 || upActive != 1 || upTotal != 2 {
		t.Fatalf("counts active=%d total=%d upActive=%d upTotal=%d, want 2,3,1,2", active, total, upActive, upTotal)
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()

	st, err := Open(filepath.Join(t.TempDir(), "meshdns.db"))
	if err != nil {
		t.Fatalf("Open = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Fatalf("Close = %v", err)
		}
	})

	return st
}

func mustCreate(t *testing.T, st *Store, server Server) {
	t.Helper()

	if err := st.CreateServer(server, "write-key-hash"); err != nil {
		t.Fatalf("CreateServer %s = %v", server.ID, err)
	}
}

func testServer(id, name string) Server {
	return Server{
		ID:           id,
		Name:         name,
		Description:  "test server",
		ServerURL:    "https://" + name + ".example.test",
		HealthURL:    "https://" + name + ".example.test/health",
		OwnerContact: name + "@example.test",
	}
}

func serverIDs(servers []Server) []string {
	ids := make([]string, 0, len(servers))
	for _, server := range servers {
		ids = append(ids, server.ID)
	}

	return ids
}
