// Manage the role mapping in Elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role-mapping.html
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

// resourceElasticsearchSecurityRoleMapping handle role mapping API call
func resourceElasticsearchSecurityRoleMapping() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchSecurityRoleMappingCreate,
		Read:   resourceElasticsearchSecurityRoleMappingRead,
		Update: resourceElasticsearchSecurityRoleMappingUpdate,
		Delete: resourceElasticsearchSecurityRoleMappingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"rules": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"roles": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
		},
	}
}

// resourceElasticsearchSecurityRoleMappingCreate  create new role mapping in Elasticsearch
func resourceElasticsearchSecurityRoleMappingCreate(d *schema.ResourceData, meta interface{}) (err error) {
	name := d.Get("name").(string)

	err = createRoleMapping(d, meta)
	if err != nil {
		return err
	}
	d.SetId(name)
	log.Infof("Created role mapping %s successfully", name)

	return resourceElasticsearchSecurityRoleMappingRead(d, meta)
}

// resourceElasticsearchSecurityRoleMappingRead read existing role mapping in Elasticsearch
func resourceElasticsearchSecurityRoleMappingRead(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()

	log.Debugf("Role mapping id:  %s", id)

	client := meta.(eshandler.ElasticsearchHandler)

	rm, err := client.RoleMappingGet(id)
	if err != nil {
		return err
	}
	if rm == nil {
		fmt.Printf("[WARN] Role mapping %s not found. Removing from state\n", id)
		log.Warnf("Role mapping %s not found. Removing from state\n", id)
		d.SetId("")
		return nil
	}

	if err = d.Set("name", id); err != nil {
		return err
	}
	if err = d.Set("enabled", rm.Enabled); err != nil {
		return err
	}
	if err = d.Set("roles", rm.Roles); err != nil {
		return err
	}
	flattenRules, err := convertInterfaceToJsonString(rm.Rules)
	if err != nil {
		return err
	}
	if err = d.Set("rules", flattenRules); err != nil {
		return err
	}
	flattenMetadata, err := convertInterfaceToJsonString(rm.Metadata)
	if err != nil {
		return err
	}
	if err = d.Set("metadata", flattenMetadata); err != nil {
		return err
	}

	log.Infof("Read role mapping %s successfully", id)
	return nil
}

// resourceElasticsearchSecurityRoleMappingUpdate update existing role mapping in Elasticsearch
func resourceElasticsearchSecurityRoleMappingUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	err = createRoleMapping(d, meta)
	if err != nil {
		return err
	}

	log.Infof("Updated role mapping %s successfully", d.Id())

	return resourceElasticsearchSecurityRoleMappingRead(d, meta)
}

// resourceElasticsearchSecurityRoleMappingDelete delete existing role mapping in Elasticsearch
func resourceElasticsearchSecurityRoleMappingDelete(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()
	log.Debugf("Role mapping id: %s", id)

	client := meta.(eshandler.ElasticsearchHandler)
	if err = client.RoleMappingDelete(id); err != nil {
		return err
	}

	d.SetId("")

	log.Infof("Deleted role mapping %s successfully", id)
	return nil

}

// createRoleMapping create or update role mapping
func createRoleMapping(d *schema.ResourceData, meta interface{}) (err error) {
	name := d.Get("name").(string)
	enabled := d.Get("enabled").(bool)
	roles := convertArrayInterfaceToArrayString(d.Get("roles").(*schema.Set).List())
	rulesStr := d.Get("rules").(string)
	metadataStr := d.Get("metadata").(string)

	client := meta.(eshandler.ElasticsearchHandler)

	rules, err := convertRawJsonTopMapString(rulesStr)
	if err != nil {
		return err
	}
	metadata, err := convertRawJsonTopMapString(metadataStr)
	if err != nil {
		return err
	}

	data := &olivere.XPackSecurityRoleMapping{
		Enabled:  enabled,
		Roles:    roles,
		Rules:    rules,
		Metadata: metadata,
	}

	if err = client.RoleMappingUpdate(name, data); err != nil {
		return err
	}

	return nil
}
