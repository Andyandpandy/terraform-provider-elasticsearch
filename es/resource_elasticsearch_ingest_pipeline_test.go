package es

import (
	"fmt"
	"testing"

	eshandler "github.com/disaster37/es-handler/v8"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchIngestPipeline(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchIngestPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchIngestPipeline,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIngestPipelineExists("elasticsearch_ingest_pipeline.test"),
				),
			},
			{
				Config: testElasticsearchIngestPipelineUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIngestPipelineExists("elasticsearch_ingest_pipeline.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_ingest_pipeline.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchIngestPipelineExists(name string) resource.TestCheckFunc {
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
		pipeline, err := client.IngestPipelineGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if pipeline == nil {
			return errors.Errorf("Ingest pipeline %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckElasticsearchIngestPipelineDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_ingest_pipeline" {
			continue
		}
		if rs.Primary.ID != "test" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		pipeline, err := client.IngestPipelineGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if pipeline != nil {
			return fmt.Errorf("Ingest pipeline %q still exists", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

var testElasticsearchIngestPipeline = `
resource "elasticsearch_ingest_pipeline" "test" {
  name 		    = "terraform-test-ingest-pipeline"
  pipeline 	  = <<EOF
{
	"description" : "My optional pipeline description",
	"processors" : [
		{
			"set" : {
				"description" : "My optional processor description",
				"field": "my-keyword-field",
				"value": "foo"
			}
		}
	]
}
EOF
}
`

var testElasticsearchIngestPipelineUpdate = `
resource "elasticsearch_ingest_pipeline" "test" {
  name 		    = "terraform-test-ingest-pipeline"
  pipeline 	  = <<EOF
{
	"description" : "My optional pipeline description",
	"processors" : [
		{
			"set" : {
				"description" : "My optional processor description",
				"field": "my-keyword-field",
				"value": "foo"
			}
		}
	]
}
EOF
}
`
