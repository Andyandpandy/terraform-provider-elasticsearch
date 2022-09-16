package es

import (
	"encoding/json"
	"reflect"
)

// optionalInterfaceJSON permit to convert string as json object
func optionalInterfaceJSON(input string) interface{} {
	if input == "" || input == "{}" {
		return nil
	}
	return json.RawMessage(input)

}

// convertArrayInterfaceToArrayString permit to convert an array of interface to an array of string
func convertArrayInterfaceToArrayString(raws []interface{}) []string {
	data := make([]string, len(raws))
	for i, raw := range raws {
		data[i] = raw.(string)
	}

	return data
}

func convertRawJsonTopMapString(raw string) (map[string]any, error) {
	result := map[string]any{}
	var err error

	if raw == "" {
		return result, nil
	}

	if err = json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, err
	}

	return result, nil
}

func convertInterfaceToJsonString(object interface{}) (string, error) {
	if object == nil {
		return "", nil
	}

	if reflect.ValueOf(object).Kind() == reflect.Map && reflect.ValueOf(object).Len() == 0 {
		return "", nil
	}

	b, err := json.Marshal(object)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
