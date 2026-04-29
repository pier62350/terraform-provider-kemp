// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccRealServerResource_basic creates a VS at 10.0.0.103, adds a real
// server, and verifies the RsIndex is assigned and tracked correctly.
func TestAccRealServerResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRealServerConfig("10.0.0.10", "8080", 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kemp_real_server.rs", "id"),
					resource.TestCheckResourceAttr("kemp_real_server.rs", "address", "10.0.0.10"),
					resource.TestCheckResourceAttr("kemp_real_server.rs", "port", "8080"),
					resource.TestCheckResourceAttr("kemp_real_server.rs", "weight", "1000"),
					resource.TestCheckResourceAttrPair(
						"kemp_real_server.rs", "virtual_service_id",
						"kemp_virtual_service.vs", "id",
					),
				),
			},
			// Update weight in-place
			{
				Config: testAccRealServerConfig("10.0.0.10", "8080", 500),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kemp_real_server.rs", "weight", "500"),
				),
			},
			// Import by "<vs_id>/<rs_id>"
			{
				ResourceName:      "kemp_real_server.rs",
				ImportState:       true,
				ImportStateIdFunc: realServerImportID("kemp_virtual_service.vs", "kemp_real_server.rs"),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRealServerConfig(rsAddr, rsPort string, weight int) string {
	return fmt.Sprintf(`
resource "kemp_virtual_service" "vs" {
  address  = "10.0.0.103"
  port     = "8080"
  protocol = "tcp"
  type     = "http"
  nickname = "tf-acc-rs-parent"
}

resource "kemp_real_server" "rs" {
  virtual_service_id = kemp_virtual_service.vs.id
  address            = %q
  port               = %q
  weight             = %d
  forward            = "nat"
}
`, rsAddr, rsPort, weight)
}

// realServerImportID returns "<vs_id>/<rs_id>".
func realServerImportID(vsAddr, rsAddr string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		vs, ok := s.RootModule().Resources[vsAddr]
		if !ok {
			return "", fmt.Errorf("resource %q not found in state", vsAddr)
		}
		rs, ok := s.RootModule().Resources[rsAddr]
		if !ok {
			return "", fmt.Errorf("resource %q not found in state", rsAddr)
		}
		return vs.Primary.ID + "/" + rs.Primary.Attributes["id"], nil
	}
}
