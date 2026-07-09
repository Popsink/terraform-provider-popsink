package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories wires the provider into the acceptance-test
// harness under the "popsink" name.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"popsink": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck fails fast when the credentials required to reach a live
// data-plane are missing. Acceptance tests only run when TF_ACC=1 (enforced by
// the testing framework); this additionally requires connectivity config.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("POPSINK_BASE_URL") == "" {
		t.Fatal("POPSINK_BASE_URL must be set for acceptance tests")
	}
	if os.Getenv("POPSINK_TOKEN") == "" {
		t.Fatal("POPSINK_TOKEN must be set for acceptance tests")
	}
}

// accProviderConfig is an empty provider block; connectivity comes from
// POPSINK_BASE_URL / POPSINK_TOKEN / POPSINK_INSECURE in the environment.
const accProviderConfig = `
provider "popsink" {}
`

// requireEnv skips the test unless the named environment variable is set,
// returning its value. Used for acceptance tests that need a pre-existing
// resource id (e.g. a datamodel, which has no create endpoint).
func requireEnv(t *testing.T, name string) string {
	t.Helper()
	v := os.Getenv(name)
	if v == "" {
		t.Skipf("%s not set; skipping (provide it to run this acceptance test)", name)
	}
	return v
}
