package settings_test

import (
	"testing"

	"github.com/gtsteffaniak/filebrowser/backend/internal/database/users"
	"github.com/gtsteffaniak/filebrowser/backend/pkg/settings"
)

func TestMergeDefaultEnabledBackendScopes_PreservesScopeQuota(t *testing.T) {
	settings.Config.Server.Sources = []*settings.Source{
		{Path: "/data", Name: "data", Config: settings.SourceConfig{DefaultEnabled: true, DefaultUserScope: "/home"}},
	}
	t.Cleanup(func() { settings.Config.Server.Sources = nil })

	perms := users.SourceFilePermissions{View: true, Download: true, Modify: true, Create: true, Configured: true}
	quota := &users.ScopeQuota{ID: "q-user-data", LimitBytes: 100 << 20, Meter: "accounted"}
	existing := []users.BackendScope{{
		Path:        "/data",
		Scope:       "/home",
		Permissions: perms,
		Quota:       quota,
	}}

	got := settings.MergeDefaultEnabledBackendScopes(existing, nil)
	if len(got) != 1 {
		t.Fatalf("len(scopes) = %d, want 1: %#v", len(got), got)
	}
	s := got[0]
	if s.Path != "/data" || s.Scope != "/home" {
		t.Fatalf("path/scope = %q %q, want %q %q", s.Path, s.Scope, "/data", "/home")
	}
	if s.Permissions != perms {
		t.Fatalf("permissions = %#v, want %#v", s.Permissions, perms)
	}
	if s.Quota == nil {
		t.Fatal("quota was dropped")
	}
	if s.Quota.ID != "q-user-data" || s.Quota.LimitBytes != 100<<20 || s.Quota.Meter != "accounted" {
		t.Fatalf("quota = %#v, want ID %q LimitBytes %d Meter %q", s.Quota, "q-user-data", 100<<20, "accounted")
	}
}

func TestMergeDefaultEnabledBackendScopes_RespectsDeclinedSources(t *testing.T) {
	settings.Config.Server.Sources = []*settings.Source{
		{Path: "/pathA", Name: "a", Config: settings.SourceConfig{DefaultEnabled: true, DefaultUserScope: "/a"}},
		{Path: "/pathB", Name: "b", Config: settings.SourceConfig{DefaultEnabled: true, DefaultUserScope: "/b"}},
	}
	t.Cleanup(func() { settings.Config.Server.Sources = nil })

	existing := []users.BackendScope{{Path: "/pathA", Scope: "/a"}}
	got := settings.MergeDefaultEnabledBackendScopes(existing, []string{"/pathB"})
	if len(got) != 1 || got[0].Path != "/pathA" {
		t.Fatalf("declined source re-added: %#v", got)
	}
}

func TestComputeDeclinedDefaultSources(t *testing.T) {
	settings.Config.Server.Sources = []*settings.Source{
		{Path: "/pathA", Name: "a", Config: settings.SourceConfig{DefaultEnabled: true}},
		{Path: "/pathB", Name: "b", Config: settings.SourceConfig{DefaultEnabled: true}},
	}
	t.Cleanup(func() { settings.Config.Server.Sources = nil })

	declined := settings.ComputeDeclinedDefaultSources([]users.BackendScope{{Path: "/pathA", Scope: "/"}})
	if len(declined) != 1 || declined[0] != "/pathB" {
		t.Fatalf("ComputeDeclinedDefaultSources() = %#v", declined)
	}
}
