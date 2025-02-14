package es

import (
	"fmt"
	"testing"

	eshandler "github.com/disaster37/es-handler/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchSnapshotRepository(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchSnapshotRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchSnapshotRepository,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSnapshotRepositoryExists("elasticsearch_snapshot_repository.test"),
				),
			},
			{
				Config: testElasticsearchSnapshotRepositoryUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSnapshotRepositoryExists("elasticsearch_snapshot_repository.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_snapshot_repository.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchSnapshotRepositoryExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No snapshot repository ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		repo, err := client.SnapshotRepositoryGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if repo == nil {
			return errors.Errorf("Repository %s nout found", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckElasticsearchSnapshotRepositoryDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_snapshot_repository" {
			continue
		}
		if rs.Primary.ID != "test" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		repo, err := client.SnapshotRepositoryGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if repo != nil {
			return fmt.Errorf("Snapshot repository %q still exists", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

var testElasticsearchSnapshotRepository = `
resource "elasticsearch_snapshot_repository" "test" {
  name		= "terraform-test"
  type 		= "fs"
  settings 	= {
	"location" =  "/tmp"
  }
}
`

var testElasticsearchSnapshotRepositoryUpdate = `
resource "elasticsearch_snapshot_repository" "test" {
  name		= "terraform-test"
  type 		= "fs"
  settings 	= {
	"location" =  "/tmp"
	"test"	= "test"
  }
}
`
