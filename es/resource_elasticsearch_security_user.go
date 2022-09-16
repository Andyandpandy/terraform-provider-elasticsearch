// Manage the user in elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role.html
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

// resourceElasticsearchSecurityUser handle the user API call
func resourceElasticsearchSecurityUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchSecurityUserCreate,
		Read:   resourceElasticsearchSecurityUserRead,
		Update: resourceElasticsearchSecurityUserUpdate,
		Delete: resourceElasticsearchSecurityUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"full_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password_hash": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"roles": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
		},
	}
}

// resourceElasticsearchSecurityUserCreate create new user in Elasticsearch
func resourceElasticsearchSecurityUserCreate(d *schema.ResourceData, meta interface{}) (err error) {
	username := d.Get("username").(string)

	err = createUser(d, meta)
	if err != nil {
		return err
	}
	d.SetId(username)

	log.Infof("Created user %s successfully", username)

	return resourceElasticsearchSecurityUserRead(d, meta)
}

// resourceElasticsearchSecurityUserRead read existing user in Elasticsearch
func resourceElasticsearchSecurityUserRead(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()

	log.Debugf("User id:  %s", id)

	client := meta.(eshandler.ElasticsearchHandler)

	user, err := client.UserGet(id)
	if err != nil {
		return err
	}
	if user == nil {
		fmt.Printf("[WARN] User %s not found - removing from state", id)
		log.Warnf("User %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	if err = d.Set("username", id); err != nil {
		return err
	}
	if err = d.Set("enabled", user.Enabled); err != nil {
		return err
	}
	if err = d.Set("email", user.Email); err != nil {
		return err
	}
	if err = d.Set("full_name", user.Fullname); err != nil {
		return err
	}
	if err = d.Set("roles", user.Roles); err != nil {
		return err
	}

	flattenMetadata, err := convertInterfaceToJsonString(user.Metadata)
	if err != nil {
		return err
	}
	if err = d.Set("metadata", flattenMetadata); err != nil {
		return err
	}

	log.Infof("Read user %s successfully", id)

	return nil
}

// resourceElasticsearchSecurityUserUpdate update existing user in Elasticsearch
func resourceElasticsearchSecurityUserUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	id := d.Id()
	enabled := d.Get("enabled").(bool)
	email := d.Get("email").(string)
	fullName := d.Get("full_name").(string)
	password := d.Get("password").(string)
	passwordHash := d.Get("password_hash").(string)
	roles := convertArrayInterfaceToArrayString(d.Get("roles").(*schema.Set).List())
	metadata := optionalInterfaceJSON(d.Get("metadata").(string))

	client := meta.(eshandler.ElasticsearchHandler)

	data := &olivere.XPackSecurityPutUserRequest{
		Enabled:  enabled,
		Email:    email,
		FullName: fullName,
		Roles:    roles,
	}
	if metadata != nil {
		data.Metadata = metadata.(map[string]interface{})
	}

	// Provide password only if it change
	if d.HasChange("password") || d.HasChange("password_hash") {
		data.Password = password
		data.PasswordHash = passwordHash
	}

	if err = client.UserUpdate(id, data); err != nil {
		return err
	}

	return resourceElasticsearchSecurityUserRead(d, meta)
}

// resourceElasticsearchSecurityUserDelete delete existing user in Elasticsearch
func resourceElasticsearchSecurityUserDelete(d *schema.ResourceData, meta interface{}) (err error) {

	id := d.Id()
	log.Debugf("User id: %s", id)

	client := meta.(eshandler.ElasticsearchHandler)

	if err = client.UserDelete(id); err != nil {
		return err
	}

	d.SetId("")

	log.Infof("Deleted user %s successfully", id)
	return nil

}

// createUser create or update user in Elasticsearch
func createUser(d *schema.ResourceData, meta interface{}) (err error) {
	username := d.Get("username").(string)
	enabled := d.Get("enabled").(bool)
	email := d.Get("email").(string)
	fullName := d.Get("full_name").(string)
	password := d.Get("password").(string)
	passwordHash := d.Get("password_hash").(string)
	roles := convertArrayInterfaceToArrayString(d.Get("roles").(*schema.Set).List())
	metadata := optionalInterfaceJSON(d.Get("metadata").(string))

	client := meta.(eshandler.ElasticsearchHandler)

	data := &olivere.XPackSecurityPutUserRequest{
		Enabled:      enabled,
		Email:        email,
		FullName:     fullName,
		Roles:        roles,
		Password:     password,
		PasswordHash: passwordHash,
	}
	if metadata != nil {
		data.Metadata = metadata.(map[string]interface{})
	}

	if err = client.UserCreate(username, data); err != nil {
		return err
	}

	return nil
}
