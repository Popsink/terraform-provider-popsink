package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/popsink/terraform-provider-popsink/internal/client"
)

// TestAccTeamDataSource looks up a team created in the same config by name.
func TestAccTeamDataSource(t *testing.T) {
	name := "tf-acc-ds-team"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
%s

resource "popsink_team" "test" {
  name        = %q
  description = "acceptance data source team"
}

data "popsink_team" "by_name" {
  name = popsink_team.test.name
}
`, accProviderConfig, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.popsink_team.by_name", "id", "popsink_team.test", "id"),
					resource.TestCheckResourceAttr("data.popsink_team.by_name", "name", name),
				),
			},
		},
	})
}

// TestAccTeamResourceDisappears verifies the provider removes a resource from
// state (rather than erroring) when it is deleted out-of-band — the 404-removal
// path. The Check deletes the team via the API, and the follow-up plan is
// expected to be non-empty (Terraform wants to recreate it).
func TestAccTeamResourceDisappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTeamConfig("tf-acc-disappear", "disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("popsink_team.test", "id"),
					testAccDeleteTeamOutOfBand("popsink_team.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDeleteTeamOutOfBand(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		c := client.NewClient(
			os.Getenv("POPSINK_BASE_URL"),
			os.Getenv("POPSINK_TOKEN"),
			os.Getenv("POPSINK_INSECURE") == "true",
		)
		return c.DeleteTeam(context.Background(), rs.Primary.ID)
	}
}
