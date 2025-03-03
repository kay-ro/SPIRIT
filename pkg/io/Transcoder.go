package io

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
)

type ConfigInformation struct {
	Parameter []ParameterInformation `json:"parameter" xml:"parameter"`
}

type ParameterInformation struct {
	Group        string `json:"group" xml:"group"`
	Name         string `json:"name" xml:"name"`
	FieldType    string `json:"type" xml:"type"`
	FieldValue   string `json:"value" xml:"value"`
	UseInFit     bool   `json:"fit" xml:"fit"`
	IsLimited    bool   `json:"limited" xml:"limited"`
	FieldMinimum string `json:"minimum" xml:"minimum"`
	FieldMaximum string `json:"maximum" xml:"maximum"`
}

func DecodeJSONFromBytes(data []byte) (*ConfigInformation, error) {
	var conf ConfigInformation
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func DecodeXMLFromBytes(data []byte) (*ConfigInformation, error) {
	var conf ConfigInformation
	if err := xml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func DecodeGOBFromBytes(data []byte) (*ConfigInformation, error) {
	var config ConfigInformation
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func EncodeJSONToBytes(config *ConfigInformation) ([]byte, error) {
	return json.Marshal(config)
}

func EncodeXMLToBytes(config *ConfigInformation) ([]byte, error) {
	return xml.Marshal(config)
}

func EncodeGOBToBytes(config *ConfigInformation) ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(config); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
