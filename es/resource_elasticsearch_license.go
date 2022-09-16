// Manage license in elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/update-license.html
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

// resourceElasticsearchLicense handle the license API call
func resourceElasticsearchLicense() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchLicenseCreate,
		Read:   resourceElasticsearchLicenseRead,
		Update: resourceElasticsearchLicenseUpdate,
		Delete: resourceElasticsearchLicenseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"license": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					var err error

					if oldValue == "" {
						oldValue = "{}"
					}
					if newValue == "" {
						newValue = "{}"
					}

					oldLicense := &olivere.XPackInfoLicense{}
					if err = json.Unmarshal([]byte(oldValue), oldLicense); err != nil {
						fmt.Printf("[ERR] Error when converting old license: %s\ndata: %s", err.Error(), oldValue)
						log.Errorf("Error when converting old license: %s\ndata: %s", err.Error(), oldValue)
					}
					newLicense := &olivere.XPackInfoLicense{}
					if err = json.Unmarshal([]byte(newValue), newLicense); err != nil {
						fmt.Printf("[ERR] Error when converting new license: %s\ndata: %s", err.Error(), oldValue)
						log.Errorf("Error when converting new license: %s\ndata: %s", err.Error(), oldValue)
					}

					return esHandler.LicenseDiff(oldLicense, newLicense)
				},
			},
			"use_basic_license": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"basic_license": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

// resourceElasticsearchLicenseCreate create license or enable basic license
func resourceElasticsearchLicenseCreate(d *schema.ResourceData, meta interface{}) (err error) {
	err = createLicense(d, meta)
	if err != nil {
		return err
	}
	d.SetId("license")
	return resourceElasticsearchLicenseRead(d, meta)
}

// resourceElasticsearchLicense update license
func resourceElasticsearchLicenseUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	err = createLicense(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchLicenseRead(d, meta)
}

// resourceElasticsearchLicenseRead read license
func resourceElasticsearchLicenseRead(d *schema.ResourceData, meta interface{}) (err error) {

	client := meta.(eshandler.ElasticsearchHandler)
	license, err := client.LicenseGet()
	if err != nil {
		return err
	}

	if license == nil {
		fmt.Printf("[WARN] License not found - removing from state")
		log.Warnf("License not found - removing from state")
		d.SetId("")
		return nil
	}

	licenseJSON, err := json.Marshal(license)
	if err != nil {
		return err
	}

	if license.Type == "basic" {
		if err = d.Set("basic_license", string(licenseJSON)); err != nil {
			return err
		}
		if err = d.Set("use_basic_license", true); err != nil {
			return err
		}
	} else {
		if err = d.Set("license", string(licenseJSON)); err != nil {
			return err
		}
		if err = d.Set("use_basic_license", false); err != nil {
			return err
		}
	}

	return nil
}

// resourceElasticsearchLicenseDelete delete license
func resourceElasticsearchLicenseDelete(d *schema.ResourceData, meta interface{}) (err error) {

	client := meta.(eshandler.ElasticsearchHandler)
	if err = client.LicenseDelete(); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

// createLicense add or update license
func createLicense(d *schema.ResourceData, meta interface{}) (err error) {
	license := d.Get("license").(string)
	useBasicLicense := d.Get("use_basic_license").(bool)

	client := meta.(eshandler.ElasticsearchHandler)

	if useBasicLicense {
		if err = client.LicenseEnableBasic(); err != nil {
			return err
		}
	} else {
		if err = client.LicenseUpdate(license); err != nil {
			return err
		}
	}

	return nil
}
