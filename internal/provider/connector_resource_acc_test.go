package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccConnectorResource covers create + import + idempotency (which exercises
// the json_configuration diff-suppression path) against a live data-plane.
func TestAccConnectorResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig("tf-acc-conn", "orders"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("popsink_connector.test", "name", "tf-acc-conn"),
					resource.TestCheckResourceAttr("popsink_connector.test", "connector_type", "KAFKA_SOURCE"),
					resource.TestCheckResourceAttrSet("popsink_connector.test", "id"),
					resource.TestCheckResourceAttrSet("popsink_connector.test", "status"),
				),
			},
			{
				// Re-applying the identical config must produce no diff; this
				// exercises Read's json_configuration diff-suppression.
				Config:   testAccConnectorConfig("tf-acc-conn", "orders"),
				PlanOnly: true,
			},
			{
				ResourceName:      "popsink_connector.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The API redacts credentials and normalizes json_configuration
				// on read, so the imported value can't match the config verbatim.
				// Computed lifecycle attributes are also not part of import.
				ImportStateVerifyIgnore: []string{"json_configuration", "desired_state", "state_timeout"},
			},
		},
	})
}

func testAccConnectorConfig(name, topic string) string {
	return fmt.Sprintf(`
%s

resource "popsink_team" "test" {
  name        = "tf-acc-conn-team"
  description = "acceptance test team"
}

resource "popsink_connector" "test" {
  name           = %q
  connector_type = "KAFKA_SOURCE"
  team_id        = popsink_team.test.id

  json_configuration = jsonencode({
    bootstrap_servers = "kafka.internal:9092"
    topic             = %q
  })
}
`, accProviderConfig, name, topic)
}
