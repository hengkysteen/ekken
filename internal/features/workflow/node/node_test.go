package node

import "testing"

func TestRegistry_RegisterDefaultsSpecVersion(t *testing.T) {
	registry := NewRegistry()
	registry.Register(NodeRegistration{
		Spec: Spec{Meta: Meta{Type: "example"}},
	})

	spec, ok := registry.GetSpec("example")
	if !ok {
		t.Fatal("expected spec to be registered")
	}
	if spec.Version != DefaultSpecVersion {
		t.Fatalf("Version = %q, want %q", spec.Version, DefaultSpecVersion)
	}
}

func TestRegistry_AllSpecsForPlatform(t *testing.T) {
	registry := NewRegistry()
	registry.Register(NodeRegistration{
		Spec: Spec{Meta: Meta{Type: "all-platforms"}},
	})
	registry.Register(NodeRegistration{
		Spec: Spec{Meta: Meta{Type: "darwin-only", Platforms: []string{"darwin"}}},
	})
	registry.Register(NodeRegistration{
		Spec: Spec{Meta: Meta{Type: "linux-only", Platforms: []string{"linux"}}},
	})

	specs := registry.AllSpecsForPlatform("darwin")
	got := make(map[string]bool, len(specs))
	for _, spec := range specs {
		got[spec.Type] = true
	}

	if !got["all-platforms"] {
		t.Fatal("expected empty platforms to support every platform")
	}
	if !got["darwin-only"] {
		t.Fatal("expected darwin node to be included on darwin")
	}
	if got["linux-only"] {
		t.Fatal("did not expect linux node to be included on darwin")
	}
}
