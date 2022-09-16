// Manage index lifecylce policy in Elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-put-lifecycle.html
// Supported version:
//  - v6
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

// resourceElasticsearchIndexLifecyclePolicy handle the index lifecycle policy API call
func resourceElasticsearchIndexLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIndexLifecyclePolicyCreate,
		Read:   resourceElasticsearchIndexLifecyclePolicyRead,
		Update: resourceElasticsearchIndexLifecyclePolicyUpdate,
		Delete: resourceElasticsearchIndexLifecyclePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"policy": {
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

					oldILM := &olivere.XPackIlmGetLifecycleResponse{}
					if err = json.Unmarshal([]byte(oldValue), oldILM); err != nil {
						fmt.Printf("[ERR] Error when converting old ILM: %s\ndata: %s", err.Error(), oldValue)
						log.Errorf("Error when converting old ILM: %s\ndata: %s", err.Error(), oldValue)
					}
					newILM := &olivere.XPackIlmGetLifecycleResponse{}
					if err = json.Unmarshal([]byte(newValue), newILM); err != nil {
						fmt.Printf("[ERR] Error when converting new ILM: %s\ndata: %s", err.Error(), oldValue)
						log.Errorf("Error when converting new ILM: %s\ndata: %s", err.Error(), oldValue)
					}

					diff, err := esHandler.ILMDiff(oldILM, newILM)
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

// resourceElasticsearchIndexLifecyclePolicyCreate create new index lifecycle policy
func resourceElasticsearchIndexLifecyclePolicyCreate(d *schema.ResourceData, meta interface{}) (err error) {
	err = createIndexLifecyclePolicy(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchIndexLifecyclePolicyRead(d, meta)
}

// resourceElasticsearchIndexLifecyclePolicyUpdate update index lifecycle policy
func resourceElasticsearchIndexLifecyclePolicyUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	err = createIndexLifecyclePolicy(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchIndexLifecyclePolicyRead(d, meta)
}

// resourceElasticsearchIndexLifecyclePolicyRead read index lifecycle policy
func resourceElasticsearchIndexLifecyclePolicyRead(d *schema.ResourceData, meta interface{}) (err error) {
	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)
	policy, err := client.ILMGet(id)
	if err != nil {
		return err
	}
	if policy == nil {
		fmt.Printf("[WARN] Index lifecycle policy %s not found - removing from state", id)
		log.Warnf("Index lifecycle policy %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	if err = d.Set("name", id); err != nil {
		return err
	}

	flattenPolicy, err := convertInterfaceToJsonString(policy)
	if err != nil {
		return err
	}
	if err = d.Set("policy", flattenPolicy); err != nil {
		return err
	}
	return nil
}

// resourceElasticsearchIndexLifecyclePolicyDelete delete index lifecycle policy
func resourceElasticsearchIndexLifecyclePolicyDelete(d *schema.ResourceData, meta interface{}) (err error) {
	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)
	if err = client.ILMDelete(id); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

// createIndexLifecyclePolicy create or update index lifecycle policy
func createIndexLifecyclePolicy(d *schema.ResourceData, meta interface{}) (err error) {
	name := d.Get("name").(string)
	policy := d.Get("policy").(string)

	data := &olivere.XPackIlmGetLifecycleResponse{}
	if err = json.Unmarshal([]byte(policy), data); err != nil {
		return err
	}

	client := meta.(eshandler.ElasticsearchHandler)
	if err = client.ILMUpdate(name, data); err != nil {
		return err
	}

	return nil
}
