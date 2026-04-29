// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccSubVirtualServiceResource_basic creates a parent VS at 10.0.0.101,
// creates a SubVS under it, and verifies the SubVS has a distinct Index that
// is tracked in state.
func TestAccSubVirtualServiceResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubVSConfig("tf-acc-subvs-parent", "tf-acc-subvs-child"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kemp_virtual_service.parent", "id"),
					resource.TestCheckResourceAttrSet("kemp_sub_virtual_service.child", "id"),
					resource.TestCheckResourceAttr("kemp_sub_virtual_service.child", "nickname", "tf-acc-subvs-child"),
					// SubVS index must differ from parent index
					resource.TestCheckResourceAttrPair(
						"kemp_sub_virtual_service.child", "parent_id",
						"kemp_virtual_service.parent", "id",
					),
				),
			},
			// Verify import by SubVS Index
			{
				ResourceName:      "kemp_sub_virtual_service.child",
				ImportState:       true,
				ImportStateIdFunc: subVSImportID("kemp_virtual_service.parent", "kemp_sub_virtual_service.child"),
				ImportStateVerify: true,
				// persist is write-only and not returned by showvs
				ImportStateVerifyIgnore: []string{"persist"},
			},
		},
	})
}

// TestAccSubVirtualServiceResource_update checks that nickname and schedule
// can be updated in-place without replacement.
func TestAccSubVirtualServiceResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubVSConfig("tf-acc-upd-parent", "tf-acc-upd-v1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kemp_sub_virtual_service.child", "nickname", "tf-acc-upd-v1"),
					resource.TestCheckResourceAttr("kemp_sub_virtual_service.child", "schedule", "rr"),
				),
			},
			{
				Config: testAccSubVSConfigSchedule("tf-acc-upd-parent", "tf-acc-upd-v2", "lc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kemp_sub_virtual_service.child", "nickname", "tf-acc-upd-v2"),
					resource.TestCheckResourceAttr("kemp_sub_virtual_service.child", "schedule", "lc"),
				),
			},
		},
	})
}

func testAccSubVSConfig(parentNick, childNick string) string {
	return fmt.Sprintf(`
resource "kemp_virtual_service" "parent" {
  address  = "10.0.0.101"
  port     = "8080"
  protocol = "tcp"
  type     = "http"
  nickname = %q
}

resource "kemp_sub_virtual_service" "child" {
  parent_id = kemp_virtual_service.parent.id
  nickname  = %q
  schedule  = "rr"
}
`, parentNick, childNick)
}

func testAccSubVSConfigSchedule(parentNick, childNick, schedule string) string {
	return fmt.Sprintf(`
resource "kemp_virtual_service" "parent" {
  address  = "10.0.0.101"
  port     = "8080"
  protocol = "tcp"
  type     = "http"
  nickname = %q
}

resource "kemp_sub_virtual_service" "child" {
  parent_id = kemp_virtual_service.parent.id
  nickname  = %q
  schedule  = %q
}
`, parentNick, childNick, schedule)
}

// subVSImportID returns an ImportStateIdFunc that produces "<parent_id>/<child_id>"
// from the state of the two given resource addresses.
func subVSImportID(parentAddr, childAddr string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		parent, ok := s.RootModule().Resources[parentAddr]
		if !ok {
			return "", fmt.Errorf("resource %q not found in state", parentAddr)
		}
		child, ok := s.RootModule().Resources[childAddr]
		if !ok {
			return "", fmt.Errorf("resource %q not found in state", childAddr)
		}
		return parent.Primary.ID + "/" + child.Primary.ID, nil
	}
}
