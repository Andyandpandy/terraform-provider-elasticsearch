// Manage component template in Elasticsearch
// API documentation:https://www.elastic.co/guide/en/elasticsearch/reference/master/indices-component-template.html
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

// resourceElasticsearchIndexComponentTemplate handle the index component template API call
func resourceElasticsearchIndexComponentTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIndexComponentTemplateCreate,
		Update: resourceElasticsearchIndexComponentTemplateUpdate,
		Read:   resourceElasticsearchIndexComponentTemplateRead,
		Delete: resourceElasticsearchIndexComponentTemplateDelete,

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

					oldComponentTemplate := &olivere.IndicesGetComponentTemplate{}
					if err = json.Unmarshal([]byte(oldValue), oldComponentTemplate); err != nil {
						fmt.Printf("[ERR] Error when converting old component template: %s\ndata: %s", err.Error(), oldValue)
						log.Errorf("Error when converting old component template: %s\ndata: %s", err.Error(), oldValue)
					}
					newComponentTemplate := &olivere.IndicesGetComponentTemplate{}
					if err = json.Unmarshal([]byte(newValue), newComponentTemplate); err != nil {
						fmt.Printf("[ERR] Error when converting new component template: %s\ndata: %s", err.Error(), oldValue)
						log.Errorf("Error when converting new component template: %s\ndata: %s", err.Error(), oldValue)
					}

					diff, err := esHandler.ComponentTemplateDiff(oldComponentTemplate, newComponentTemplate)
					if err != nil {
						fmt.Printf("[ERR] Error when diff component template: %s", err.Error())
						log.Errorf("Error when diff component template: %s", err.Error())
					}

					return diff == ""
				},
			},
		},
	}
}

// resourceElasticsearchIndexComponentTemplateCreate create index component template
func resourceElasticsearchIndexComponentTemplateCreate(d *schema.ResourceData, meta interface{}) (err error) {

	err = createIndexComponentTemplate(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchIndexComponentTemplateRead(d, meta)
}

// resourceElasticsearchIndexComponentTemplateUpdate update index component template
func resourceElasticsearchIndexComponentTemplateUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	err = createIndexComponentTemplate(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchIndexComponentTemplateRead(d, meta)
}

// resourceElasticsearchIndexComponentTemplateRead read index component template
func resourceElasticsearchIndexComponentTemplateRead(d *schema.ResourceData, meta interface{}) (err error) {
	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)
	ct, err := client.ComponentTemplateGet(id)
	if err != nil {
		return err
	}
	if ct == nil {
		fmt.Printf("[WARN] Index component template %s not found - removing from state", id)
		log.Warnf("Index component template %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	indexComponentTemplateJSON, err := json.Marshal(ct)
	if err != nil {
		return err
	}

	log.Debugf("Get index component template %s successfully:%+v", id, string(indexComponentTemplateJSON))
	if err = d.Set("name", d.Id()); err != nil {
		return err
	}
	if err = d.Set("template", string(indexComponentTemplateJSON)); err != nil {
		return err
	}
	return nil
}

// resourceElasticsearchIndexComponentTemplateDelete delete index template
func resourceElasticsearchIndexComponentTemplateDelete(d *schema.ResourceData, meta interface{}) (err error) {
	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)
	if err = client.ComponentTemplateDelete(id); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

// createIndexComponentTemplate create or update index component template
func createIndexComponentTemplate(d *schema.ResourceData, meta interface{}) (err error) {
	name := d.Get("name").(string)
	template := d.Get("template").(string)
	client := meta.(eshandler.ElasticsearchHandler)

	data := &olivere.IndicesGetComponentTemplate{}
	if err = json.Unmarshal([]byte(template), data); err != nil {
		return err
	}

	if err = client.ComponentTemplateUpdate(name, data); err != nil {
		return err
	}

	return nil
}
