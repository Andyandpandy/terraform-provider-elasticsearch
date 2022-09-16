package es

import (
	"fmt"
	"testing"

	eshandler "github.com/disaster37/es-handler/v8"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchTransform(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchTransformDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchTransform,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchTransformExists("elasticsearch_transform.test"),
				),
			},
			{
				Config: testElasticsearchTransformUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchTransformExists("elasticsearch_transform.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_transform.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchTransformExists(name string) resource.TestCheckFunc {
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
		transform, err := client.TransformGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if transform == nil {
			return errors.Errorf("Transform %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckElasticsearchTransformDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_transform" {
			continue
		}
		if rs.Primary.ID != "test" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		transform, err := client.TransformGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if transform != nil {
			return fmt.Errorf("Transform %q still exists", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

var testElasticsearchTransform = `
resource "elasticsearch_transform" "test" {
  name 		= "terraform-test-transform"
  transform 	= <<EOF
{
	"source": {
		"index": ["ecs-*"],
		"query": {
			"term": {
				"geoip.continent_name": {
					"value": "Asia"
				}
			}
		}
	},
	"pivot": {
		"group_by": {
			"customer_id": {
				"terms": {
					"field": "customer_id"
				}
			}
		},
		"aggregations": {
			"max_price": {
				"max": {
					"field": "taxful_total_price"
				}
			}
		}
	},
	"description": "Maximum priced ecommerce data by customer_id in Asia",
	"dest": {
		"index": "kibana_sample_data_ecommerce_transform1"
	},
	"frequency": "5m",
	"sync": {
		"time": {
			"field": "order_date",
			"delay": "60s"
		}
	},
	"retention_policy": {
		"time": {
			"field": "order_date",
			"max_age": "30d"
		}
	}
}
EOF
}
`

var testElasticsearchTransformUpdate = `
resource "elasticsearch_transform" "test" {
  name 		= "terraform-test-transform"
  transform 	= <<EOF
{
	"source": {
		"index": ["ecs-*"],
		"query": {
			"term": {
				"geoip.continent_name": {
					"value": "Asia"
				}
			}
		}
	},
	"pivot": {
		"group_by": {
			"customer_id": {
				"terms": {
					"field": "customer_id"
				}
			}
		},
		"aggregations": {
			"max_price": {
				"max": {
					"field": "taxful_total_price"
				}
			}
		}
	},
	"description": "Maximum priced ecommerce data by customer_id in Asia",
	"dest": {
		"index": "kibana_sample_data_ecommerce_transform1"
	},
	"frequency": "5m",
	"sync": {
		"time": {
			"field": "order_date",
			"delay": "60s"
		}
	},
	"retention_policy": {
		"time": {
			"field": "order_date",
			"max_age": "30d"
		}
	}
}
EOF
}
`
