// Manage ingest pipeline in Elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-apis.html
// Supported version:
//  - v7

package es

import (
	"encoding/json"
	"fmt"

	eshandler "github.com/disaster37/es-handler/v8"
	olivere "github.com/olivere/elastic/v7"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	log "github.com/sirupsen/logrus"
)

// resourceElasticsearchIngestPipeline handle the ingest pipeline API call
func resourceElasticsearchIngestPipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIngestPipelineCreate,
		Update: resourceElasticsearchIngestPipelineUpdate,
		Read:   resourceElasticsearchIngestPipelineRead,
		Delete: resourceElasticsearchIngestPipelineDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"pipeline": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: diffSuppressIngestPipeline,
			},
		},
	}
}

// resourceElasticsearchIngestPipelineCreate create ingest pipeline
func resourceElasticsearchIngestPipelineCreate(d *schema.ResourceData, meta interface{}) (err error) {

	err = createIngestPipeline(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchIngestPipelineRead(d, meta)
}

// resourceElasticsearchIngestPipelineUpdate update ingest pipeline
func resourceElasticsearchIngestPipelineUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	err = createIngestPipeline(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchIngestPipelineRead(d, meta)
}

// resourceElasticsearchIngestPipelineRead read ingest pipeline
func resourceElasticsearchIngestPipelineRead(d *schema.ResourceData, meta interface{}) (err error) {
	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)
	pipeline, err := client.IngestPipelineGet(id)
	if err != nil {
		return err
	}

	if pipeline == nil {
		fmt.Printf("[WARN] Ingest pipeline %s not found - removing from state", id)
		log.Warnf("Ingest pipeline %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	pipelineJSON, err := json.Marshal(pipeline)
	if err != nil {
		return err
	}

	log.Debugf("Get ingest pipeline %s successfully:%+v", id, string(pipelineJSON))
	if err = d.Set("name", d.Id()); err != nil {
		return err
	}
	if err = d.Set("pipeline", string(pipelineJSON)); err != nil {
		return err
	}
	return nil
}

// resourceElasticsearchIngestPipelineDelete delete ingest pipeline
func resourceElasticsearchIngestPipelineDelete(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)
	if err = client.IngestPipelineDelete(id); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

// createIngestPipeline create or update ingest pipeline
func createIngestPipeline(d *schema.ResourceData, meta interface{}) (err error) {
	name := d.Get("name").(string)
	pipeline := d.Get("pipeline").(string)

	client := meta.(eshandler.ElasticsearchHandler)

	data := &olivere.IngestGetPipeline{}
	if err = json.Unmarshal([]byte(pipeline), data); err != nil {
		return err
	}
	if err = client.IngestPipelineUpdate(name, data); err != nil {
		return err
	}

	return nil
}
