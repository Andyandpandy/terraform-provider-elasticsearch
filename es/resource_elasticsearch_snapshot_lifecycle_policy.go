// Manage snapshot lifecycle policy in elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/slm-api-put.html
// Supported version:
//  - v7

package es

import (
	"encoding/json"
	"fmt"

	eshandler "github.com/disaster37/es-handler/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	log "github.com/sirupsen/logrus"
)

// resourceElasticsearchSnapshotLifecyclePolicy handle the snapshot lifecycle policy API call
func resourceElasticsearchSnapshotLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchSnapshotLifecyclePolicyCreate,
		Read:   resourceElasticsearchSnapshotLifecyclePolicyRead,
		Update: resourceElasticsearchSnapshotLifecyclePolicyUpdate,
		Delete: resourceElasticsearchSnapshotLifecyclePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"snapshot_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"schedule": {
				Type:     schema.TypeString,
				Required: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"configs": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					return suppressEquivalentJSONWithExclude(k, oldValue, newValue, d, map[string]any{
						"ignore_unavailable":   false,
						"include_global_state": false,
					})
				},
			},
			"retention": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
		},
	}
}

// resourceElasticsearchSnapshotLifecyclePolicyCreate create snapshot lifecycle policy
func resourceElasticsearchSnapshotLifecyclePolicyCreate(d *schema.ResourceData, meta interface{}) (err error) {

	name := d.Get("name").(string)

	err = createSnapshotLifecyclePolicy(d, meta)
	if err != nil {
		return err
	}
	d.SetId(name)
	return resourceElasticsearchSnapshotLifecyclePolicyRead(d, meta)
}

// resourceElasticsearchSnapshotLifecyclePolicyUpdate update snapshot lifecycle policy
func resourceElasticsearchSnapshotLifecyclePolicyUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	err = createSnapshotLifecyclePolicy(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchSnapshotLifecyclePolicyRead(d, meta)
}

// resourceElasticsearchSnapshotLifecyclePolicyRead read snapshot lifecycle policy
func resourceElasticsearchSnapshotLifecyclePolicyRead(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)
	policy, err := client.SLMGet(id)
	if err != nil {
		return err
	}

	if policy == nil {
		fmt.Printf("[WARN] Snapshot lifecycle policy %s not found - removing from state", id)
		log.Warnf("Snapshot lifecycle policy %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	if err = d.Set("name", id); err != nil {
		return err
	}
	if err = d.Set("snapshot_name", policy.Name); err != nil {
		return err
	}
	if err = d.Set("schedule", policy.Schedule); err != nil {
		return err
	}
	if err = d.Set("repository", policy.Repository); err != nil {
		return err
	}

	flattenConfigs, err := convertInterfaceToJsonString(policy.Config)
	if err != nil {
		return err
	}
	if err = d.Set("configs", flattenConfigs); err != nil {
		return err
	}

	flattenRetention, err := convertInterfaceToJsonString(policy.Retention)
	if err != nil {
		return err
	}
	if err = d.Set("retention", flattenRetention); err != nil {
		return err
	}

	return nil
}

// resourceElasticsearchSnapshotLifecyclePolicyDelete delete snapshot lifecycle policy
func resourceElasticsearchSnapshotLifecyclePolicyDelete(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)
	if err = client.SLMDelete(id); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

// createSnapshotLifecyclePolicy permit to create or update snapshot lifecycle policy
func createSnapshotLifecyclePolicy(d *schema.ResourceData, meta interface{}) (err error) {
	name := d.Get("name").(string)
	snapshotName := d.Get("snapshot_name").(string)
	schedule := d.Get("schedule").(string)
	repository := d.Get("repository").(string)
	configStr := d.Get("configs").(string)
	retentionStr := d.Get("retention").(string)

	client := meta.(eshandler.ElasticsearchHandler)

	config := &eshandler.ElasticsearchSLMConfig{}
	if err = json.Unmarshal([]byte(configStr), config); err != nil {
		return err
	}

	retention := &eshandler.ElasticsearchSLMRetention{}
	if err = json.Unmarshal([]byte(retentionStr), retention); err != nil {
		return err
	}

	data := &eshandler.SnapshotLifecyclePolicySpec{
		Name:       snapshotName,
		Schedule:   schedule,
		Repository: repository,
		Config:     *config,
		Retention:  retention,
	}

	if err = client.SLMUpdate(name, data); err != nil {
		return err
	}

	return nil
}
