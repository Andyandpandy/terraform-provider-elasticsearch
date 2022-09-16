// Manage snapshot repository in elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/modules-snapshots.html
// Supported version:
//  - v6
//  - v7

package es

import (
	"fmt"

	eshandler "github.com/disaster37/es-handler/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	olivere "github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

// resourceElasticsearchSnapshotRepository handle the snapshot repository API call
func resourceElasticsearchSnapshotRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchSnapshotRepositoryCreate,
		Read:   resourceElasticsearchSnapshotRepositoryRead,
		Update: resourceElasticsearchSnapshotRepositoryUpdate,
		Delete: resourceElasticsearchSnapshotRepositoryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"settings": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// resourceElasticsearchSnapshotRepositoryCreate create snapshot repository
func resourceElasticsearchSnapshotRepositoryCreate(d *schema.ResourceData, meta interface{}) (err error) {

	name := d.Get("name").(string)

	err = createSnapshotRepository(d, meta)
	if err != nil {
		return err
	}
	d.SetId(name)
	return resourceElasticsearchSnapshotRepositoryRead(d, meta)
}

// resourceElasticsearchSnapshotRepositoryUpdate update the snapshot repository
func resourceElasticsearchSnapshotRepositoryUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	err = createSnapshotRepository(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchSnapshotRepositoryRead(d, meta)
}

// resourceElasticsearchSnapshotRepositoryRead read the sanpshot repository
func resourceElasticsearchSnapshotRepositoryRead(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)

	repo, err := client.SnapshotRepositoryGet(id)
	if err != nil {
		return err
	}

	if repo == nil {
		fmt.Printf("[WARN] Snapshot repository %s not found - removing from state", id)
		log.Warnf("Snapshot repository %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	if err = d.Set("name", id); err != nil {
		return err
	}
	if err = d.Set("type", repo.Type); err != nil {
		return err
	}
	if err = d.Set("settings", repo.Settings); err != nil {
		return err
	}

	return nil
}

// resourceElasticsearchSnapshotRepositoryDelete delete the snapshot repository
func resourceElasticsearchSnapshotRepositoryDelete(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()

	client := meta.(eshandler.ElasticsearchHandler)
	if err = client.SnapshotRepositoryDelete(id); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

// createSnapshotRepository create or update snapshot repository
func createSnapshotRepository(d *schema.ResourceData, meta interface{}) (err error) {
	name := d.Get("name").(string)
	snapshotType := d.Get("type").(string)
	settings := d.Get("settings").(map[string]interface{})

	client := meta.(eshandler.ElasticsearchHandler)

	data := &olivere.SnapshotRepositoryMetaData{
		Type:     snapshotType,
		Settings: settings,
	}

	if err = client.SnapshotRepositoryUpdate(name, data); err != nil {
		return err
	}

	return nil
}
