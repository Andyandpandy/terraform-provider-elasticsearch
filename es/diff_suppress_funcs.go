package es

import (
	"encoding/json"
	"fmt"
	"reflect"

	eshandler "github.com/disaster37/es-handler/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

// suppressEquivalentJSON permit to compare state store as JSON string
func suppressEquivalentJSON(k, old, new string, d *schema.ResourceData) bool {

	var err error
	oldObj := map[string]any{}
	newObj := map[string]any{}

	if old == "" {
		old = "{}"
	}
	if new == "" {
		new = "{}"
	}

	if err = json.Unmarshal([]byte(old), &oldObj); err != nil {
		fmt.Printf("[ERR] Error when converting current Json: %s\ndata: %s", err.Error(), old)
		log.Errorf("Error when converting current Json: %s\ndata: %s", err.Error(), old)
	}
	if err = json.Unmarshal([]byte(new), &newObj); err != nil {
		fmt.Printf("[ERR] Error when converting current Json: %s\ndata: %s", err.Error(), new)
		log.Errorf("Error when converting current Json: %s\ndata: %s", err.Error(), new)
	}

	diff, err := eshandler.StandardDiff(oldObj, newObj, logEntry, nil)
	if err != nil {
		fmt.Printf("[ERR] Error when diff JSON: %s", err.Error())
		log.Errorf("Error when diff Json: %s", err.Error())
	}

	return diff == ""
}

func suppressEquivalentJSONWithExclude(k, old, new string, d *schema.ResourceData, exclude map[string]any) bool {

	var err error
	oldObj := map[string]any{}
	newObj := map[string]any{}

	if old == "" {
		old = "{}"
	}
	if new == "" {
		new = "{}"
	}

	if err = json.Unmarshal([]byte(old), &oldObj); err != nil {
		fmt.Printf("[ERR] Error when converting current Json: %s\ndata: %s", err.Error(), old)
		log.Errorf("Error when converting current Json: %s\ndata: %s", err.Error(), old)
	}
	if err = json.Unmarshal([]byte(new), &newObj); err != nil {
		fmt.Printf("[ERR] Error when converting current Json: %s\ndata: %s", err.Error(), new)
		log.Errorf("Error when converting current Json: %s\ndata: %s", err.Error(), new)
	}

	diff, err := eshandler.StandardDiff(oldObj, newObj, logEntry, exclude)
	if err != nil {
		fmt.Printf("[ERR] Error when diff JSON: %s", err.Error())
		log.Errorf("Error when diff Json: %s", err.Error())
	}

	return diff == ""
}

// diffSuppressTransform permit to compare transform in current state vs from API
func diffSuppressTransform(k, old, new string, d *schema.ResourceData) bool {
	oo := &Transform{}
	no := &Transform{}

	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		fmt.Printf("[ERR] Error when converting to Transform on old object: %s", err.Error())
		log.Errorf("Error when converting to Transform on old object: %s\n%s", err.Error(), old)
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		fmt.Printf("[ERR] Error when converting to Transform on new object: %s", err.Error())
		log.Errorf("Error when converting to Transform on new object: %s\n%s", err.Error(), new)
		return false
	}

	oo.Id = ""
	oo.CreateTime = 0
	oo.Version = ""

	return reflect.DeepEqual(no, oo)
}

// diffSuppressIngestPipeline permit to compare ingest pipeline in current state vs from API
func diffSuppressIngestPipeline(k, old, new string, d *schema.ResourceData) bool {
	oo := &elastic.IngestGetPipeline{}
	no := &elastic.IngestGetPipeline{}

	if err := json.Unmarshal([]byte(old), &oo); err != nil {
		fmt.Printf("[ERR] Error when converting to IngestGetPipeline on old object: %s", err.Error())
		log.Errorf("Error when converting to IngestGetPipeline on old object: %s\n%s", err.Error(), old)
		return false
	}
	if err := json.Unmarshal([]byte(new), &no); err != nil {
		fmt.Printf("[ERR] Error when converting to IngestGetPipeline on new object: %s", err.Error())
		log.Errorf("Error when converting to IngestGetPipeline on new object: %s\n%s", err.Error(), new)
		return false
	}

	return reflect.DeepEqual(no, oo)
}
