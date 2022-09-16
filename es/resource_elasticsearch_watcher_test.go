package es

import (
	"fmt"
	"testing"

	eshandler "github.com/disaster37/es-handler/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchWatcher(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchWatcherDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchWatcher,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchWatcherExists("elasticsearch_watcher.test"),
				),
			},
			{
				Config: testElasticsearchWatcherUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchWatcherExists("elasticsearch_watcher.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_watcher.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchWatcherExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No watcher ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		watcher, err := client.WatchGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if watcher == nil {
			return errors.Errorf("Watcher %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckElasticsearchWatcherDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_watcher" {
			continue
		}
		if rs.Primary.ID != "test" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(eshandler.ElasticsearchHandler)
		watcher, err := client.WatchGet(rs.Primary.ID)
		if err != nil {
			return err
		}
		if watcher != nil {
			return fmt.Errorf("Watcher %q still exists", rs.Primary.ID)
		}

		return nil

	}

	return nil
}

var testElasticsearchWatcher = `
resource "elasticsearch_watcher" "test" {
  name		= "terraform-test"
  trigger	= <<EOF
{
	"schedule" : { "cron" : "0 0/1 * * * ?" }
}
EOF
  input		= <<EOF
{
	"search" : {
      "request" : {
        "indices" : [
          "logstash*"
        ],
		"search_type": "query_then_fetch",
		"rest_total_hits_as_int": true,
        "body" : {
          "query" : {
            "bool" : {
              "must" : {
                "match": {
                   "response": 404
                }
              },
              "filter" : {
                "range": {
                  "@timestamp": {
                    "from": "{{ctx.trigger.scheduled_time}}||-5m",
                    "to": "{{ctx.trigger.triggered_time}}"
                  }
                }
              }
            }
          }
        }
      }
    }
}
EOF
  condition		= <<EOF
{
	"compare" : { "ctx.payload.hits.total" : { "gt" : 0 }}
}
EOF
  actions		= <<EOF
{
	"email_admin" : {
      "email" : {
		"profile": "standard",
        "to" : ["admin@domain.host.com"],
        "subject" : "404 recently encountered"
      }
    }
}
EOF
}
`

var testElasticsearchWatcherUpdate = `
resource "elasticsearch_watcher" "test" {
  name		= "terraform-test"
  trigger	= <<EOF
{
	"schedule" : { "cron" : "1 0/1 * * * ?" }
}
EOF
  input		= <<EOF
{
	"search" : {
      "request" : {
        "indices" : [
          "logstash*"
        ],
		"search_type": "query_then_fetch",
		"rest_total_hits_as_int": true,
        "body" : {
          "query" : {
            "bool" : {
              "must" : {
                "match": {
                   "response": 404
                }
              },
              "filter" : {
                "range": {
                  "@timestamp": {
                    "from": "{{ctx.trigger.scheduled_time}}||-5m",
                    "to": "{{ctx.trigger.triggered_time}}"
                  }
                }
              }
            }
          }
        }
      }
    }
}
EOF
  condition		= <<EOF
{
	"compare" : { "ctx.payload.hits.total" : { "gt" : 0 }}
}
EOF
  actions		= <<EOF
{
	"email_admin" : {
      "email" : {
		"profile" : "standard",
        "to" : ["admin@domain.host.com"],
        "subject" : "404 recently encountered"
      }
    }
}
EOF
}
`
