package es

import (
	"fmt"
	"testing"

	eshandler "github.com/disaster37/es-handler/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchSecurityRoleMapping(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchSecurityRoleMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchSecurityRoleMapping,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSecurityRoleMappingExists("elasticsearch_role_mapping.test"),
				),
			},
			{
				Config: testElasticsearchSecurityRoleMappingUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSecurityRoleMappingExists("elasticsearch_role_mapping.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_role_mapping.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchSecurityRoleMappingExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No role mapping ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		rm, err := client.RoleMappingGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if rm == nil {
			return errors.Errorf("Role mapping %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckElasticsearchSecurityRoleMappingDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_role_mapping" {
			continue
		}
		if rs.Primary.ID != "test" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		rm, err := client.RoleMappingGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if rm != nil {
			return fmt.Errorf("Role mapping %q still exists", rs.Primary.ID)
		}

		return nil

	}

	return nil
}

var testElasticsearchSecurityRoleMapping = `
resource "elasticsearch_role_mapping" "test" {
  name 		= "terraform-test"
  enabled 	= "true"
  roles 	= ["superuser"]
  rules 	= <<EOF
{
	"field": {
		"groups": "cn=admins,dc=example,dc=com"
	}
}
EOF
}
`

var testElasticsearchSecurityRoleMappingUpdate = `
resource "elasticsearch_role_mapping" "test" {
  name 		= "terraform-test"
  enabled 	= "true"
  roles 	= ["superuser"]
  rules 	= <<EOF
{
	"field": {
		"groups": "cn=admins2,dc=example,dc=com"
	}
}
EOF
}
`
