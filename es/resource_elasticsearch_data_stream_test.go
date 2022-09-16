package es

import (
	"context"
	"fmt"
	"testing"

	eshandler "github.com/disaster37/es-handler/v8"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchDataStream(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchDataStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchDataStream,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchDataStreamExists("elasticsearch_data_stream.test"),
				),
			},
			{
				Config: testElasticsearchDataStreamUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchDataStreamExists("elasticsearch_data_stream.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_data_stream.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchDataStreamExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No inde ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler).Client()
		res, err := client.API.Indices.GetDataStream(
			client.API.Indices.GetDataStream.WithName(rs.Primary.ID),
			client.API.Indices.GetDataStream.WithContext(context.Background()),
			client.API.Indices.GetDataStream.WithPretty(),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			return errors.Errorf("Error when get data stream %s: %s", rs.Primary.ID, res.String())
		}

		return nil
	}
}

func testCheckElasticsearchDataStreamDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_data_stream" {
			continue
		}

		if rs.Primary.ID != "test" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler).Client()
		res, err := client.API.Indices.GetDataStream(
			client.API.Indices.GetDataStream.WithName(rs.Primary.ID),
			client.API.Indices.GetDataStream.WithContext(context.Background()),
			client.API.Indices.GetDataStream.WithPretty(),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			if res.StatusCode == 404 {
				return nil
			}
		}

		return fmt.Errorf("Data stream %q still exists", rs.Primary.ID)
	}

	return nil
}

var testElasticsearchDataStream = `
resource "elasticsearch_index_template" "test-data-stream" {
  name 		= "test-data-stream"
  template 	= <<EOF
{
	"index_patterns": ["terraform-test"],
	"data_stream": {},
	"priority": 2
}
EOF
}

resource "elasticsearch_data_stream" "test" {
  name 		= "terraform-test"

	depends_on = [ elasticsearch_index_template.test-data-stream ]
}
`

var testElasticsearchDataStreamUpdate = `
resource "elasticsearch_index_template" "test-data-stream" {
  name 		= "test-data-stream"
  template 	= <<EOF
{
	"index_patterns": ["terraform-test"],
	"data_stream": {},
	"priority": 2
}
EOF
}

resource "elasticsearch_data_stream" "test" {
  name 		= "terraform-test"

	depends_on = [ elasticsearch_index_template.test-data-stream ]
}
`
