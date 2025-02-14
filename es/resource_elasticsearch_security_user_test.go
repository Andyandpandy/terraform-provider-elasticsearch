package es

import (
	"fmt"
	"testing"

	eshandler "github.com/disaster37/es-handler/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchSecurityUser(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchSecurityUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchSecurityUser,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSecurityUserExists("elasticsearch_user.test"),
				),
			},
			{
				Config: testElasticsearchSecurityUserUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSecurityUserExists("elasticsearch_user.test"),
				),
			},
			{
				ResourceName:            "elasticsearch_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "password_hash"},
			},
		},
	})
}

func testCheckElasticsearchSecurityUserExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No user ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		user, err := client.UserGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if user == nil {
			return errors.Errorf("User %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckElasticsearchSecurityUserDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_user" {
			continue
		}
		if rs.Primary.ID != "test" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		user, err := client.UserGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if user != nil {
			return fmt.Errorf("User %q still exists", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

var testElasticsearchSecurityUser = `
resource "elasticsearch_user" "test" {
  username 	= "terraform-test"
  enabled 	= "true"
  email 	= "no@no.no"
  full_name = "test"
  password 	= "changeme"
  roles 	= ["kibana_user"]
}
`

var testElasticsearchSecurityUserUpdate = `
resource "elasticsearch_user" "test" {
  username 	= "terraform-test"
  enabled 	= "true"
  email 	= "no@no.no"
  full_name = "test2"
  password 	= "changeme2"
  roles 	= ["kibana_user"]
}
`
