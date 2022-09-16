package es

import (
	"fmt"
	"testing"

	eshandler "github.com/disaster37/es-handler/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchLicense(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchLicenseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchLicense,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchLicenseExists("elasticsearch_license.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_license.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchLicenseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No license ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		license, err := client.LicenseGet()
		if err != nil {
			return err
		}
		if license == nil {
			return errors.New("License not found")
		}

		return nil
	}
}

func testCheckElasticsearchLicenseDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_license" {
			continue
		}
		if rs.Primary.ID != "test" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		//license, err := client.LicenseGet()
		_, err := client.LicenseGet()
		if err != nil {
			return err
		}
		/*
			if license != nil {
				// The basic license continue to exist
				return errors.New("License still exist")
			}
		*/

		return nil
	}

	return nil
}

var testElasticsearchLicense = `
resource "elasticsearch_license" "test" {
  use_basic_license = "true"
}
`
