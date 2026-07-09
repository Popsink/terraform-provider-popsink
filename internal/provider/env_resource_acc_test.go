package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccEnvResource covers create + import for an environment with a plaintext
// broker configuration.
func TestAccEnvResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvConfig("tf-acc-env"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("popsink_env.test", "name", "tf-acc-env"),
					resource.TestCheckResourceAttrSet("popsink_env.test", "id"),
					resource.TestCheckResourceAttr("popsink_env.test", "retention_configuration.bootstrap_server", "kafka.internal:9092"),
				),
			},
			{
				ResourceName:      "popsink_env.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Credentials are stripped from the read shape, so they can't be
				// verified against the imported state.
				ImportStateVerifyIgnore: []string{
					"retention_configuration.sasl_username",
					"retention_configuration.sasl_password",
					"retention_configuration.ca_cert",
					"retention_configuration.cert",
					"retention_configuration.key",
				},
			},
		},
	})
}

func testAccEnvConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "popsink_env" "test" {
  name = %q

  retention_configuration = {
    bootstrap_server = "kafka.internal:9092"
  }
}
`, accProviderConfig, name)
}
