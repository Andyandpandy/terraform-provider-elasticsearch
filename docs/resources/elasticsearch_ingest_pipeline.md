# elasticsearch_ingest_pipeline Resource Source

This resource permit to manage the ingest pipeline in Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-apis.html

***Supported Elasticsearch version:***
  - v8

## Example Usage

It will create ingest pipeline.

```tf
resource elasticsearch_ingest_pipeline "test" {
  name 		= "terraform-test"
  pipeline 	= <<EOF
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
```

## Argument Reference

***The following arguments are supported:***
  - **name**: (required) Identifier for the ingest pipeline.
  - **pipeline**: (required) The pipeline specification. It's a string as JSON object.

## Attribute Reference

NA