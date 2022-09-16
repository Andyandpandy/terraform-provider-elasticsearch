# elasticsearch_data_stream

This resource permit to manage the index data stream in Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/data-stream-apis.html

***Supported Elasticsearch version:***
  - v8

## Example Usage

It will create data stream index.

> You need to have index template with `data_sream` that match your index before to create data stream index.

```tf
resource elasticsearch_data_stream "test" {
  name 		= "terraform-test"
}
```

## Argument Reference

***The following arguments are supported:***
  - **name**: (required) The data stream index name.

## Attribute Reference

NA