// Manage index template in Elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/index-templates.html
// Supported version:
//  - v7

package es

import (
	"encoding/json"
	"fmt"

	eshandler "github.com/disaster37/es-handler/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	olivere "github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

// resourceElasticsearchIndexTemplate handle the index template API call
func resourceElasticsearchIndexTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIndexTemplateCreate,
		Update: resourceElasticsearchIndexTemplateUpdate,
		Read:   resourceElasticsearchIndexTemplateRead,
		Delete: resourceElasticsearchIndexTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"template": {
				Type:     schema.TypeString,
				Required: true,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					var err error

					if oldValue == "" {
						oldValue = "{}"
					}
					if newValue == "" {
						newValue = "{}"
					}

					oldIndexTemplate := &olivere.IndicesGetIndexTemplate{}
					if err = json.Unmarshal([]byte(oldValue), oldIndexTemplate); err != nil {
						fmt.Printf("[ERR] Error when converting old index template: %s\ndata: %s", err.Error(), oldValue)
						log.Errorf("Error when converting old index template: %s\ndata: %s", err.Error(), oldValue)
					}
					newIndexTemplate := &olivere.IndicesGetIndexTemplate{}
					if err = json.Unmarshal([]byte(newValue), newIndexTemplate); err != nil {
						fmt.Printf("[ERR] Error when converting new index template: %s\ndata: %s", err.Error(), oldValue)
						log.Errorf("Error when converting new index template: %s\ndata: %s", err.Error(), oldValue)
					}

					diff, err := esHandler.IndexTemplateDiff(oldIndexTemplate, newIndexTemplate)
					if err != nil {
						fmt.Printf("[ERR] Error when diff index template: %s", err.Error())
						log.Errorf("Error when diff index template: %s", err.Error())
					}

					return diff == ""
				},
			},
		},
	}
}

// resourceElasticsearchIndexTemplateCreate create index template
func resourceElasticsearchIndexTemplateCreate(d *schema.ResourceData, meta interface{}) (err error) {

	err = createIndexTemplate(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchIndexTemplateRead(d, meta)
}

// resourceElasticsearchIndexTemplateUpdate update index template
func resourceElasticsearchIndexTemplateUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	err = createIndexTemplate(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchIndexTemplateRead(d, meta)
}

// resourceElasticsearchIndexTemplateRead read index template
func resourceElasticsearchIndexTemplateRead(d *schema.ResourceData, meta interface{}) (err error) {
	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)
	it, err := client.IndexTemplateGet(id)
	if err != nil {
		return err
	}

	if it == nil {
		fmt.Printf("[WARN] Index template %s not found - removing from state", id)
		log.Warnf("Index template %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	indexTemplateJSON, err := json.Marshal(it)
	if err != nil {
		return err
	}

	log.Debugf("Get index template %s successfully:%+v", id, string(indexTemplateJSON))
	if err = d.Set("name", d.Id()); err != nil {
		return err
	}
	if err = d.Set("template", string(indexTemplateJSON)); err != nil {
		return err
	}
	return nil
}

// resourceElasticsearchIndexTemplateDelete delete index template
func resourceElasticsearchIndexTemplateDelete(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)

	if err = client.IndexTemplateDelete(id); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

// createIndexTemplate create or update index template
func createIndexTemplate(d *schema.ResourceData, meta interface{}) (err error) {
	name := d.Get("name").(string)
	template := d.Get("template").(string)

	client := meta.(eshandler.ElasticsearchHandler)

	data := &olivere.IndicesGetIndexTemplate{}
	if err = json.Unmarshal([]byte(template), data); err != nil {
		return err
	}
	if err = client.IndexTemplateUpdate(name, data); err != nil {
		return err
	}

	return nil
}
