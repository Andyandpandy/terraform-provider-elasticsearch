package es

import (
	"fmt"
	"testing"

	eshandler "github.com/disaster37/es-handler/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchSecurityRole(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchSecurityRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchSecurityRole,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSecurityRoleExists("elasticsearch_role.test"),
				),
			},
			{
				Config: testElasticsearchSecurityRoleUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSecurityRoleExists("elasticsearch_role.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchSecurityRoleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No role ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		role, err := client.RoleGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if role == nil {
			return errors.Errorf("Role %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckElasticsearchSecurityRoleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_role" {
			continue
		}
		if rs.Primary.ID != "test" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		role, err := client.RoleGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if role != nil {
			return fmt.Errorf("Security role %q still exists", rs.Primary.ID)
		}

		return nil

	}

	return nil
}

var testElasticsearchSecurityRole = `
resource "elasticsearch_role" "test" {
  name = "terraform-test"
  indices {
	  names = ["logstash-*"]
	  privileges = ["read"]
  }
  indices {
	  names = ["app-*"]
	  privileges = ["read"]
  }
  cluster = ["all"]
}
`

var testElasticsearchSecurityRoleUpdate = `
resource "elasticsearch_role" "test" {
  name = "terraform-test"
  indices {
	  names = ["logstash-*"]
	  privileges = ["write"]
  }
  indices {
	  names = ["app-*"]
	  privileges = ["read"]
  }
  cluster = ["all"]
}
`
