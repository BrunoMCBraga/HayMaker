package haymakerutil

import (
	"encoding/json"
	"errors"
)

func ConvertStringToJSONStruct(stringJson *string) (interface{}, error) {

	var configStruct interface{}
	unmarshallErr := json.Unmarshal([]byte(*stringJson), &configStruct)

	if unmarshallErr != nil {
		return nil, errors.New("|" + "util->json_util->ConvertStringToJSONStruct:" + unmarshallErr.Error() + "|")
	}

	return configStruct, unmarshallErr

}

func ConvertStructToJSONString(jsonStructure interface{}) (*string, error) {

	marshalResult, marshalError := json.MarshalIndent(jsonStructure, "", "    ")

	if marshalError != nil {
		return nil, errors.New("|" + "util->json_util->ConvertStructToJSONString:" + marshalError.Error() + "|")
	}

	jsonString := string(marshalResult)

	return &jsonString, nil

}
