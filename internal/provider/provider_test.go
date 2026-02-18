package provider

import (
	"testing"
)

func TestNew(t *testing.T) {
	version := "1.0.0"
	providerFunc := New(version)

	if providerFunc == nil {
		t.Fatal("expected provider function, got nil")
	}

	provider := providerFunc()
	if provider == nil {
		t.Fatal("expected provider instance, got nil")
	}

	popsinkProvider, ok := provider.(*popsinkProvider)
	if !ok {
		t.Fatal("expected *popsinkProvider type")
	}

	if popsinkProvider.version != version {
		t.Errorf("expected version %s, got %s", version, popsinkProvider.version)
	}
}

func TestProviderMetadata(t *testing.T) {
	// This is a basic test to ensure the provider can be instantiated
	provider := New("test")()
	if provider == nil {
		t.Fatal("expected provider, got nil")
	}
}
