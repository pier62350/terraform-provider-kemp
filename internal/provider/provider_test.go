// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories is the provider factory used by all
// acceptance tests. It builds an in-process provider with the test version.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"kemp": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck fails the test fast if the live-LoadMaster credentials
// aren't present. Acceptance tests only run when TF_ACC=1, so this guards
// against cryptic plugin errors in CI.
func testAccPreCheck(t *testing.T) {
	if os.Getenv("KEMP_HOST") == "" {
		t.Fatal("KEMP_HOST must be set for acceptance tests")
	}
	if os.Getenv("KEMP_API_KEY") == "" && (os.Getenv("KEMP_USERNAME") == "" || os.Getenv("KEMP_PASSWORD") == "") {
		t.Fatal("KEMP_API_KEY or KEMP_USERNAME+KEMP_PASSWORD must be set for acceptance tests")
	}
}

// Example acceptance test — creates a fresh virtual service at 10.0.0.100.
// Acceptance tests must NOT collide with any pre-existing virtual services
// on the LoadMaster they target.
func TestAccVirtualServiceResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "kemp_virtual_service" "test" {
  address  = "10.0.0.100"
  port     = "8080"
  protocol = "tcp"
  type     = "http"
  nickname = "tf-acc-test"
  enabled  = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kemp_virtual_service.test", "address", "10.0.0.100"),
					resource.TestCheckResourceAttr("kemp_virtual_service.test", "port", "8080"),
					resource.TestCheckResourceAttr("kemp_virtual_service.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("kemp_virtual_service.test", "nickname", "tf-acc-test"),
					resource.TestCheckResourceAttrSet("kemp_virtual_service.test", "id"),
				),
			},
		},
	})
}
