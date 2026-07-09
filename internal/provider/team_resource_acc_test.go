package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccTeamResource covers create, update (in place), import, and read-back.
func TestAccTeamResource(t *testing.T) {
	name := "tf-acc-team"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTeamConfig(name, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("popsink_team.test", "name", name),
					resource.TestCheckResourceAttr("popsink_team.test", "description", "initial description"),
					resource.TestCheckResourceAttrSet("popsink_team.test", "id"),
				),
			},
			{
				ResourceName:      "popsink_team.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTeamConfig(name, "updated description"),
				Check:  resource.TestCheckResourceAttr("popsink_team.test", "description", "updated description"),
			},
		},
	})
}

func testAccTeamConfig(name, description string) string {
	return fmt.Sprintf(`
%s

resource "popsink_team" "test" {
  name        = %q
  description = %q
}
`, accProviderConfig, name, description)
}
