# elasticsearch_transform Resource Source

This resource permit to manage the transform in Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/transform-apis.html

***Supported Elasticsearch version:***
  - v8

## Example Usage

It will create transform.

```tf
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
```

## Argument Reference

***The following arguments are supported:***
  - **name**: (required) Identifier for the transform.
  - **transform**: (required) The transform specification. It's a string as JSON object.

## Attribute Reference

NA