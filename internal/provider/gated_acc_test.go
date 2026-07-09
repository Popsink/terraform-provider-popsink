package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccTeamMemberResource requires POPSINK_TEST_USER_ID (an existing user to
// add to a freshly created team); it is skipped otherwise.
func TestAccTeamMemberResource(t *testing.T) {
	userID := requireEnv(t, "POPSINK_TEST_USER_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
%s

resource "popsink_team" "test" {
  name        = "tf-acc-member-team"
  description = "acceptance membership team"
}

resource "popsink_team_member" "test" {
  team_id = popsink_team.test.id
  user_id = %q
  role    = "member"
}
`, accProviderConfig, userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("popsink_team_member.test", "user_id", userID),
					resource.TestCheckResourceAttr("popsink_team_member.test", "role", "member"),
					resource.TestCheckResourceAttrSet("popsink_team_member.test", "id"),
				),
			},
			{
				ResourceName:            "popsink_team_member.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"role"},
				// Import ID is the composite "team_id/member_id".
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["popsink_team_member.test"]
					if !ok {
						return "", fmt.Errorf("resource popsink_team_member.test not found in state")
					}
					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["team_id"], rs.Primary.ID), nil
				},
			},
		},
	})
}

// TestAccSubscriptionResource requires a pre-existing datamodel
// (POPSINK_TEST_DATAMODEL_ID) and target connector
// (POPSINK_TEST_TARGET_CONNECTOR_ID); datamodels have no create endpoint.
func TestAccSubscriptionResource(t *testing.T) {
	datamodelID := requireEnv(t, "POPSINK_TEST_DATAMODEL_ID")
	targetID := requireEnv(t, "POPSINK_TEST_TARGET_CONNECTOR_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
%s

resource "popsink_subscription" "test" {
  name                = "tf-acc-subscription"
  datamodel_id        = %q
  target_connector_id = %q
  desired_state       = "paused"
}
`, accProviderConfig, datamodelID, targetID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("popsink_subscription.test", "name", "tf-acc-subscription"),
					resource.TestCheckResourceAttr("popsink_subscription.test", "desired_state", "paused"),
					resource.TestCheckResourceAttrSet("popsink_subscription.test", "config_hash"),
				),
			},
			{
				ResourceName:            "popsink_subscription.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"smt_config"},
			},
		},
	})
}
