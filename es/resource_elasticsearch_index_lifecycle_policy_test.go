package es

import (
	"fmt"
	"testing"

	eshandler "github.com/disaster37/es-handler/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchIndexLifecyclePolicy(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchIndexLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchIndexLifecyclePolicy,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIndexLifecyclePolicyExists("elasticsearch_index_lifecycle_policy.test"),
				),
			},
			{
				Config: testElasticsearchIndexLifecyclePolicyUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIndexLifecyclePolicyExists("elasticsearch_index_lifecycle_policy.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_index_lifecycle_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchIndexLifecyclePolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No index lifecycle policy ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		policy, err := client.ILMGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if policy == nil {
			return errors.Errorf("ILM %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckElasticsearchIndexLifecyclePolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_index_lifecycle_policy" {
			continue
		}
		if rs.Primary.ID != "test" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		policy, err := client.ILMGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if policy != nil {
			return fmt.Errorf("Index lifecycle policy %q still exists", rs.Primary.ID)
		}

		return nil

	}

	return nil
}

var testElasticsearchIndexLifecyclePolicy = `
resource "elasticsearch_index_lifecycle_policy" "test" {
  name = "terraform-test"
  policy = <<EOF
{
  "policy": {
    "phases": {
      "warm": {
        "min_age": "10d",
        "actions": {
          "forcemerge": {
            "max_num_segments": 1
          }
        }
      },
      "delete": {
        "min_age": "30d",
        "actions": {
          "delete": {
			"delete_searchable_snapshot": true
		  }
        }
      }
    }
  }
}
EOF
}
`

var testElasticsearchIndexLifecyclePolicyUpdate = `
resource "elasticsearch_index_lifecycle_policy" "test" {
  name = "terraform-test"
  policy = <<EOF
{
  "policy": {
    "phases": {
      "warm": {
        "min_age": "10d",
        "actions": {
          "forcemerge": {
            "max_num_segments": 1
          }
        }
      },
      "delete": {
        "min_age": "31d",
        "actions": {
          "delete": {
			"delete_searchable_snapshot": true
		  }
        }
      }
    }
  }
}
EOF
}
`
