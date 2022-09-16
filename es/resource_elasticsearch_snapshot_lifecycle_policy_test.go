package es

import (
	"fmt"
	"testing"

	eshandler "github.com/disaster37/es-handler/v8"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchSnapshotLifecyclePolicy(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchSnapshotLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchSnapshotLifecyclePolicy,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSnapshotLifecyclePolicyExists("elasticsearch_snapshot_lifecycle_policy.test"),
				),
			},
			{
				Config: testElasticsearchSnapshotLifecyclePolicyUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSnapshotLifecyclePolicyExists("elasticsearch_snapshot_lifecycle_policy.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_snapshot_lifecycle_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchSnapshotLifecyclePolicyExists(name string) resource.TestCheckFunc {
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
		policy, err := client.SLMGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if policy == nil {
			return errors.Errorf("SLM policy %s not found", rs.Primary.ID)
		}

		// Manage Bug https://github.com/elastic/elasticsearch/issues/47664

		return nil
	}
}

func testCheckElasticsearchSnapshotLifecyclePolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_snapshot_lifecycle_policy" {
			continue
		}
		if rs.Primary.ID != "test" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		policy, err := client.SLMGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if policy != nil {
			return fmt.Errorf("Snapshot lifecycle policy %q still exists", rs.Primary.ID)
		}

		return nil

	}

	return nil
}

var testElasticsearchSnapshotLifecyclePolicy = `

resource "elasticsearch_snapshot_repository" "test" {
  name		= "test"
  type 		= "fs"
  settings 	= {
	"location" =  "/tmp"
  }
}

resource "elasticsearch_snapshot_lifecycle_policy" "test" {
  name			= "terraform-test"
  snapshot_name = "<daily-snap-{now/d}>"
  schedule 		= "0 30 1 * * ?"
  repository    = "${elasticsearch_snapshot_repository.test.name}"
  configs		= <<EOF
{
	"indices": ["test-*"],
	"ignore_unavailable": false,
	"include_global_state": false
}
EOF
  retention     = <<EOF
{
    "expire_after": "7d",
    "min_count": 5,
    "max_count": 10
} 
EOF
}
`

var testElasticsearchSnapshotLifecyclePolicyUpdate = `

resource "elasticsearch_snapshot_repository" "test" {
  name		= "test"
  type 		= "fs"
  settings 	= {
	"location" =  "/tmp"
  }
}

resource "elasticsearch_snapshot_lifecycle_policy" "test" {
  name			= "terraform-test"
  snapshot_name = "<daily-snap-{now/d}>"
  schedule 		= "1 30 1 * * ?"
  repository    = "${elasticsearch_snapshot_repository.test.name}"
  configs		= <<EOF
{
	"indices": ["test-*"],
	"ignore_unavailable": false,
	"include_global_state": false
}
EOF
  retention     = <<EOF
{
    "expire_after": "7d",
    "min_count": 5,
    "max_count": 10
} 
EOF
}
`
