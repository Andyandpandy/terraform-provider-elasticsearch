// Manage the role in elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role.html
// Supported version:
//  - v6
//  - v7

package es

import (
	"encoding/json"
	"fmt"
	"reflect"

	eshandler "github.com/disaster37/es-handler/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	log "github.com/sirupsen/logrus"
)

// resourceElasticsearchSecurityRole handle the role API call
func resourceElasticsearchSecurityRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchSecurityRoleCreate,
		Read:   resourceElasticsearchSecurityRoleRead,
		Update: resourceElasticsearchSecurityRoleUpdate,
		Delete: resourceElasticsearchSecurityRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"run_as": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"global": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"indices": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"names": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"privileges": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"query": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressEquivalentJSON,
						},
						"field_security": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressEquivalentJSON,
						},
					},
				},
			},
			"applications": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"application": {
							Type:     schema.TypeString,
							Required: true,
						},
						"privileges": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"resources": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

// resourceElasticsearchSecurityRoleCreate create new role in Elasticsearch
func resourceElasticsearchSecurityRoleCreate(d *schema.ResourceData, meta interface{}) (err error) {
	name := d.Get("name").(string)

	err = createRole(d, meta)
	if err != nil {
		return err
	}
	d.SetId(name)

	log.Infof("Created role %s successfully", name)

	return resourceElasticsearchSecurityRoleRead(d, meta)
}

// resourceElasticsearchSecurityRoleRead read existing role in Elasticsearch
func resourceElasticsearchSecurityRoleRead(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()

	log.Debugf("Role id:  %s", id)

	client := meta.(eshandler.ElasticsearchHandler)
	role, err := client.RoleGet(id)

	if err != nil {
		return err
	}
	if role == nil {
		fmt.Printf("[WARN] Role %s not found - removing from state", id)
		log.Warnf("Role %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	if err = d.Set("name", id); err != nil {
		return err
	}

	flattenIndices, err := flattenIndicesMapping(role.Indices)
	if err != nil {
		return err
	}
	if err = d.Set("indices", flattenIndices); err != nil {
		return fmt.Errorf("error setting indices: %w", err)
	}
	if err = d.Set("cluster", role.Cluster); err != nil {
		return fmt.Errorf("error setting cluster: %w", err)
	}

	if err = d.Set("applications", flattenApplicationsMapping(role.Applications)); err != nil {
		return fmt.Errorf("error setting applications: %w", err)
	}

	global := ""
	if len(role.Global) > 0 {
		globalB, err := json.Marshal(role.Global)
		if err != nil {
			return err
		}
		global = string(globalB)
	}
	if err = d.Set("global", global); err != nil {
		return err
	}
	if err = d.Set("run_as", role.RunAs); err != nil {
		return err
	}

	flattenMetdata, err := convertInterfaceToJsonString(role.Metadata)
	if err != nil {
		return err
	}
	if err = d.Set("metadata", flattenMetdata); err != nil {
		return err
	}

	log.Infof("Read role %s successfully", id)

	return nil
}

// resourceElasticsearchSecurityRoleUpdate update existing role in Elasticsearch
func resourceElasticsearchSecurityRoleUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	err = createRole(d, meta)
	if err != nil {
		return err
	}

	log.Infof("Updated role %s successfully", d.Id())

	return resourceElasticsearchSecurityRoleRead(d, meta)
}

// resourceElasticsearchSecurityRoleDelete delete existing role in Elasticsearch
func resourceElasticsearchSecurityRoleDelete(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()
	log.Debugf("Role id: %s", id)

	client := meta.(eshandler.ElasticsearchHandler)
	if err = client.RoleDelete(id); err != nil {
		return err
	}

	d.SetId("")

	log.Infof("Deleted role %s successfully", id)
	return nil

}

// createRole create or update role in Elasticsearch
func createRole(d *schema.ResourceData, meta interface{}) (err error) {
	name := d.Get("name").(string)
	indices := buildRolesIndicesPermissions(d.Get("indices").(*schema.Set).List())
	applications := buildRolesApplicationPrivileges(d.Get("applications").(*schema.Set).List())
	cluster := convertArrayInterfaceToArrayString(d.Get("cluster").(*schema.Set).List())
	global := optionalInterfaceJSON(d.Get("global").(string))
	runAs := convertArrayInterfaceToArrayString(d.Get("run_as").(*schema.Set).List())
	metadata := optionalInterfaceJSON(d.Get("metadata").(string))

	client := meta.(eshandler.ElasticsearchHandler)

	data := &eshandler.XPackSecurityRole{
		Cluster:      cluster,
		Applications: applications,
		Indices:      indices,
		RunAs:        runAs,
	}

	if global != nil {
		data.Global = global.(map[string]interface{})
	}
	if metadata != nil {
		data.Metadata = metadata.(map[string]interface{})
	}

	if err = client.RoleUpdate(name, data); err != nil {
		return err
	}

	return nil
}

// buildRolesIndicesPermissions convert list to list of RoleIndicesPermissions objects
func buildRolesIndicesPermissions(raws []interface{}) []eshandler.XPackSecurityIndicesPermissions {

	rolesIndicesPermissions := make([]eshandler.XPackSecurityIndicesPermissions, 0, len(raws))

	for _, raw := range raws {
		m := raw.(map[string]interface{})
		// Mitigeate bug https://github.com/hashicorp/terraform-plugin-sdk/issues/895
		if len(m["names"].(*schema.Set).List()) == 0 {
			continue
		}
		roleIndicesPermisions := eshandler.XPackSecurityIndicesPermissions{
			Names:         convertArrayInterfaceToArrayString(m["names"].(*schema.Set).List()),
			Privileges:    convertArrayInterfaceToArrayString(m["privileges"].(*schema.Set).List()),
			Query:         m["query"].(string),
			FieldSecurity: optionalInterfaceJSON(m["field_security"].(string)),
		}

		rolesIndicesPermissions = append(rolesIndicesPermissions, roleIndicesPermisions)

	}

	return rolesIndicesPermissions
}

// buildRolesApplicationPrivileges convert list to list of RoleApplicationPrivileges objects
func buildRolesApplicationPrivileges(raws []interface{}) []eshandler.XPackSecurityApplicationPrivileges {
	rolesApplicationPrivileges := make([]eshandler.XPackSecurityApplicationPrivileges, 0, len(raws))

	for _, raw := range raws {
		m := raw.(map[string]interface{})

		// Mitigeate bug https://github.com/hashicorp/terraform-plugin-sdk/issues/895
		if m["application"].(string) == "" {
			continue
		}
		roleApplicationPrivileges := eshandler.XPackSecurityApplicationPrivileges{
			Application: m["application"].(string),
			Privileges:  convertArrayInterfaceToArrayString(m["privileges"].(*schema.Set).List()),
			Resources:   convertArrayInterfaceToArrayString(m["resources"].(*schema.Set).List()),
		}

		rolesApplicationPrivileges = append(rolesApplicationPrivileges, roleApplicationPrivileges)

	}

	return rolesApplicationPrivileges
}

func flattenIndiceMapping(indice eshandler.XPackSecurityIndicesPermissions) (map[string]interface{}, error) {
	if reflect.ValueOf(indice).IsZero() {
		return nil, nil
	}

	tfMap := make(map[string]interface{})
	tfMap["names"] = indice.Names
	tfMap["privileges"] = indice.Privileges

	if indice.Query != "" {
		queryB, err := json.Marshal(indice.Query)
		if err != nil {
			return nil, err
		}
		tfMap["query"] = string(queryB)
	}

	if indice.FieldSecurity != nil {
		fiedlSecurityB, err := json.Marshal(indice.FieldSecurity)
		if err != nil {
			return nil, err
		}
		tfMap["field_security"] = string(fiedlSecurityB)
	}

	return tfMap, nil
}

func flattenIndicesMapping(indices []eshandler.XPackSecurityIndicesPermissions) ([]interface{}, error) {
	if indices == nil {
		return nil, nil
	}

	tfList := make([]interface{}, 0, len(indices))

	for _, indice := range indices {
		flattenIndices, err := flattenIndiceMapping(indice)
		if err != nil {
			return nil, err
		}
		tfList = append(tfList, flattenIndices)
	}

	return tfList, nil

}

func flattenApplicationMapping(application eshandler.XPackSecurityApplicationPrivileges) map[string]interface{} {
	if reflect.ValueOf(application).IsZero() {
		return nil
	}

	tfMap := make(map[string]interface{})
	tfMap["application"] = application.Application
	tfMap["privileges"] = application.Privileges
	tfMap["resources"] = application.Resources

	return tfMap
}

func flattenApplicationsMapping(applications []eshandler.XPackSecurityApplicationPrivileges) []interface{} {
	if applications == nil {
		return nil
	}

	tfList := make([]interface{}, 0, len(applications))
	for _, application := range applications {
		tfList = append(tfList, flattenApplicationMapping(application))
	}

	return tfList
}
