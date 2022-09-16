// Manage transform in Elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-apis.html
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

type TransformGetResponse struct {
	Transforms []Transform `json:"transforms,omitempty"`
}

type Transform struct {
	Id          string            `json:"id"`
	Version     string            `json:"version"`
	CreateTime  int64             `json:"create_time"`
	Source      TransformSource   `json:"source"`
	Dest        TransformDest     `json:"dest"`
	Frequency   string            `json:"frequency"`
	Sync        TransformSync     `json:"sync"`
	Pivot       TransformPivot    `json:"pivot"`
	Description string            `json:"description"`
	Settings    TransformSettings `json:"settings"`
}

type TransformSource struct {
	Index []string    `json:"index"`
	Query interface{} `json:"query,omitempty"`
}

type TransformDest struct {
	Index    string `json:"index"`
	Pipeline string `json:"pipeline"`
}

type TransformSync struct {
	Time TransformTime `json:"time"`
}

type TransformTime struct {
	Field string `json:"field"`
	Delay string `json:"delay"`
}

type TransformPivot struct {
	GroupBy      map[string]interface{} `json:"group_by"`
	Aggregations map[string]interface{} `json:"aggregations"`
}

type TransformSettings struct {
	MaxPageSearchSize int `json:"max_page_search_size"`
}

// resourceElasticsearchTransform handle the transform API call
func resourceElasticsearchTransform() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchTransformCreate,
		Read:   resourceElasticsearchTransformRead,
		Delete: resourceElasticsearchTransformDelete,
		Update: resourceElasticsearchTransformUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"transform": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: diffSuppressTransform,
			},
		},
	}
}

// resourceElasticsearchTransformCreate create transform
func resourceElasticsearchTransformCreate(d *schema.ResourceData, meta interface{}) (err error) {

	err = createTransform(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchTransformRead(d, meta)
}

// resourceElasticsearchTransformUpdate update transform
func resourceElasticsearchTransformUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	err = createTransform(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchTransformRead(d, meta)
}

// resourceElasticsearchTransformRead read transform
func resourceElasticsearchTransformRead(d *schema.ResourceData, meta interface{}) (err error) {
	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)
	transform, err := client.TransformGet(id)
	if err != nil {
		return err
	}

	if transform == nil {
		fmt.Printf("[WARN] Transform %s not found - removing from state", id)
		log.Warnf("Transform %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	transformJSON, err := json.Marshal(transform)
	if err != nil {
		return err
	}

	log.Debugf("Get transform %s successfully:%+v", id, string(transformJSON))
	if err = d.Set("name", d.Id()); err != nil {
		return err
	}
	if err = d.Set("transform", string(transformJSON)); err != nil {
		return err
	}
	return nil
}

// resourceElasticsearchTransformDelete delete transform
func resourceElasticsearchTransformDelete(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)

	if err = client.TransformDelete(id); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

// createTransform create or update transform
func createTransform(d *schema.ResourceData, meta interface{}) (err error) {
	name := d.Get("name").(string)
	transform := d.Get("transform").(string)

	client := meta.(eshandler.ElasticsearchHandler)

	data := &eshandler.Transform{}
	if err = json.Unmarshal([]byte(transform), data); err != nil {
		return err
	}
	if err = client.TransformUpdate(name, data); err != nil {
		return err
	}

	return nil
}
