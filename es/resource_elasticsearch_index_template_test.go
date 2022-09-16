package es

import (
	"fmt"
	"testing"

	eshandler "github.com/disaster37/es-handler/v8"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchIndexTemplate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchIndexTemplate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIndexTemplateExists("elasticsearch_index_template.test"),
				),
			},
			{
				Config: testElasticsearchIndexTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIndexTemplateExists("elasticsearch_index_template.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_index_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchIndexTemplateExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No index ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		it, err := client.IndexTemplateGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if it == nil {
			return errors.Errorf("Index template %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckElasticsearchIndexTemplateDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_index_template" {
			continue
		}
		if rs.Primary.ID != "test" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		it, err := client.IndexTemplateGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if it != nil {
			return fmt.Errorf("Index template %q still exists", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

var testElasticsearchIndexTemplate = `
resource "elasticsearch_index_template" "test" {
  name 		= "terraform-test-index-template"
  template 	= <<EOF
{
	"index_patterns": ["test-index-template"],
	"template": {
		"settings": {
			"index.refresh_interval": "5s",
			"index.lifecycle.name": "policy-logstash-backup",
    		"index.lifecycle.rollover_alias": "logstash-backup-alias"
		}
	},
	"priority": 2
}
EOF
}
`

var testElasticsearchIndexTemplateUpdate = `
resource "elasticsearch_index_template" "test" {
  name 		= "terraform-test-index-template"
  template 	= <<EOF
{
	"index_patterns": ["test-index-template"],
	"template": {
		"settings": {
			"index.refresh_interval": "3s",
			"index.lifecycle.name": "policy-logstash-backup",
    		"index.lifecycle.rollover_alias": "logstash-backup-alias"
		}
	},
	"priority": 2
}
EOF
}
`
