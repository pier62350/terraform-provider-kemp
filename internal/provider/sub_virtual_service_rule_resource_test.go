// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccSubVirtualServiceRuleResource_basic creates a parent VS + SubVS +
// match-content rule, attaches the rule to the SubVS, and verifies state.
func TestAccSubVirtualServiceRuleResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubVSRuleConfig("tf-acc-svsr-rule"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kemp_virtual_service.parent", "id"),
					resource.TestCheckResourceAttrSet("kemp_sub_virtual_service.child", "id"),
					resource.TestCheckResourceAttr("kemp_match_content_rule.r", "name", "tf-acc-svsr-rule"),
					resource.TestCheckResourceAttr("kemp_sub_virtual_service_rule.attachment", "rule", "tf-acc-svsr-rule"),
					resource.TestCheckResourceAttrPair(
						"kemp_sub_virtual_service_rule.attachment", "sub_virtual_service_id",
						"kemp_sub_virtual_service.child", "id",
					),
					resource.TestCheckResourceAttrPair(
						"kemp_sub_virtual_service_rule.attachment", "parent_virtual_service_id",
						"kemp_virtual_service.parent", "id",
					),
				),
			},
			// Import by "<parent_vs_id>/<sub_vs_id>/<rule_name>"
			{
				ResourceName:      "kemp_sub_virtual_service_rule.attachment",
				ImportState:       true,
				ImportStateIdFunc: subVSRuleImportID("kemp_virtual_service.parent", "kemp_sub_virtual_service.child", "tf-acc-svsr-rule"),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSubVSRuleConfig(ruleName string) string {
	return fmt.Sprintf(`
resource "kemp_virtual_service" "parent" {
  address  = "10.0.0.102"
  port     = "8080"
  protocol = "tcp"
  type     = "http"
  nickname = "tf-acc-svsr-parent"
}

resource "kemp_sub_virtual_service" "child" {
  parent_id = kemp_virtual_service.parent.id
  nickname  = "tf-acc-svsr-child"
}

resource "kemp_match_content_rule" "r" {
  name    = %q
  pattern = "^/api/"
}

resource "kemp_sub_virtual_service_rule" "attachment" {
  parent_virtual_service_id = kemp_virtual_service.parent.id
  sub_virtual_service_id    = kemp_sub_virtual_service.child.id
  rule                      = kemp_match_content_rule.r.name
}
`, ruleName)
}

// subVSRuleImportID returns "<parent_id>/<child_id>/<ruleName>".
func subVSRuleImportID(parentAddr, childAddr, ruleName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		parent, ok := s.RootModule().Resources[parentAddr]
		if !ok {
			return "", fmt.Errorf("resource %q not found in state", parentAddr)
		}
		child, ok := s.RootModule().Resources[childAddr]
		if !ok {
			return "", fmt.Errorf("resource %q not found in state", childAddr)
		}
		return parent.Primary.ID + "/" + child.Primary.ID + "/" + ruleName, nil
	}
}
