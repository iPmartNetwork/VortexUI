package domain

import "testing"

func TestNodeMultiCoreHelpers(t *testing.T) {
	n := &Node{Core: CoreXray, EnabledCores: []CoreType{CoreXray, CoreSingbox}}
	if !n.IsMultiCore() {
		t.Fatal("expected multi-core")
	}
	if n.SyncCoreType() != CoreMulti {
		t.Fatalf("SyncCoreType = %q, want multi", n.SyncCoreType())
	}
	if n.ResolveInboundCore("") != CoreXray {
		t.Fatalf("default core = %q", n.ResolveInboundCore(""))
	}
	if n.ResolveInboundCore(CoreSingbox) != CoreSingbox {
		t.Fatalf("override core = %q", n.ResolveInboundCore(CoreSingbox))
	}
	if !n.CoreEnabled(CoreSingbox) {
		t.Fatal("singbox should be enabled")
	}
	if n.CoreEnabled(CoreMulti) {
		t.Fatal("multi is not an engine")
	}
}

func TestNodeLegacySingleCore(t *testing.T) {
	n := &Node{Core: CoreSingbox}
	if n.IsMultiCore() {
		t.Fatal("legacy node should be single-core")
	}
	if got := n.NormalizedEnabledCores(); len(got) != 1 || got[0] != CoreSingbox {
		t.Fatalf("NormalizedEnabledCores = %v", got)
	}
}
