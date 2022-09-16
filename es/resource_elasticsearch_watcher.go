// Manage the watcher in elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-put-watch.html
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

// resourceElasticsearchWatcher handle the watcher API call
func resourceElasticsearchWatcher() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchWatcherCreate,
		Read:   resourceElasticsearchWatcherRead,
		Update: resourceElasticsearchWatcherUpdate,
		Delete: resourceElasticsearchWatcherDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"trigger": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"input": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"condition": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"actions": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"throttle_period": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
		},
	}
}

// resourceElasticsearchWatcherCreate create new watcher in Elasticsearch
func resourceElasticsearchWatcherCreate(d *schema.ResourceData, meta interface{}) (err error) {
	name := d.Get("name").(string)

	err = createWatcher(d, meta)
	if err != nil {
		return err
	}
	d.SetId(name)

	log.Infof("Created watcher %s successfully", name)

	return resourceElasticsearchWatcherRead(d, meta)
}

// resourceElasticsearchWatcherRead read existing watch in Elasticsearch
func resourceElasticsearchWatcherRead(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()

	log.Debugf("Watcher id:  %s", id)

	client := meta.(eshandler.ElasticsearchHandler)
	watcher, err := client.WatchGet(id)
	if err != nil {
		return err
	}
	if watcher == nil {
		fmt.Printf("[WARN] Watcher %s not found - removing from state", id)
		log.Warnf("Watcher %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	if err = d.Set("name", id); err != nil {
		return err
	}

	flattenTrigger, err := convertInterfaceToJsonString(watcher.Trigger)
	if err != nil {
		return err
	}
	if err = d.Set("trigger", flattenTrigger); err != nil {
		return err
	}

	flattenInput, err := convertInterfaceToJsonString(watcher.Input)
	if err != nil {
		return err
	}
	if err = d.Set("input", flattenInput); err != nil {
		return err
	}

	flattenCondition, err := convertInterfaceToJsonString(watcher.Condition)
	if err != nil {
		return err
	}
	if err = d.Set("condition", flattenCondition); err != nil {
		return err
	}

	flattenActions, err := convertInterfaceToJsonString(watcher.Actions)
	if err != nil {
		return err
	}
	if err = d.Set("actions", flattenActions); err != nil {
		return err
	}

	flattenMetadata, err := convertInterfaceToJsonString(watcher.Metadata)
	if err != nil {
		return err
	}
	if err = d.Set("metadata", flattenMetadata); err != nil {
		return err
	}

	if watcher.ThrottlePeriod != "" {
		if err = d.Set("throttle_period", watcher.ThrottlePeriod); err != nil {
			return err
		}
	}

	log.Infof("Read watcher %s successfully", id)

	return nil
}

// resourceElasticsearchWatcherUpdate update existing watcher in Elasticsearch
func resourceElasticsearchWatcherUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	err = createWatcher(d, meta)
	if err != nil {
		return err
	}

	log.Infof("Updated watcher %s successfully", d.Id())

	return resourceElasticsearchWatcherRead(d, meta)
}

// resourceElasticsearchWatcherDelete delete existing watcher in Elasticsearch
func resourceElasticsearchWatcherDelete(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()
	log.Debugf("Watcher id: %s", id)

	client := meta.(eshandler.ElasticsearchHandler)

	if err = client.WatchDelete(id); err != nil {
		return err
	}

	d.SetId("")

	log.Infof("Deleted watcher %s successfully", id)
	return nil

}

// createWatcher create or update watcher in Elasticsearch
func createWatcher(d *schema.ResourceData, meta interface{}) (err error) {
	name := d.Get("name").(string)
	triggerStr := d.Get("trigger").(string)
	inputStr := d.Get("input").(string)
	conditionStr := d.Get("condition").(string)
	actionStr := d.Get("actions").(string)
	metadata := optionalInterfaceJSON(d.Get("metadata").(string))
	throttlePeriod := d.Get("throttle_period").(string)

	trigger := new(map[string]map[string]interface{})
	if err = json.Unmarshal([]byte(triggerStr), trigger); err != nil {
		return err
	}
	input := new(map[string]map[string]interface{})
	if err = json.Unmarshal([]byte(inputStr), input); err != nil {
		return err
	}
	condition := new(map[string]map[string]interface{})
	if err = json.Unmarshal([]byte(conditionStr), condition); err != nil {
		return err
	}
	action := new(map[string]map[string]interface{})
	if err = json.Unmarshal([]byte(actionStr), action); err != nil {
		return err
	}

	client := meta.(eshandler.ElasticsearchHandler)

	data := &olivere.XPackWatch{
		Trigger:        *trigger,
		Input:          *input,
		Condition:      *condition,
		Actions:        *action,
		ThrottlePeriod: throttlePeriod,
	}
	if metadata != nil {
		data.Metadata = metadata.(map[string]interface{})
	}

	if err = client.WatchUpdate(name, data); err != nil {
		return err
	}

	return nil
}
