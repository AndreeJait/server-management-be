package entity

import (
	"database/sql/driver"
	"encoding/json"
)

type StringMap map[string]string

func (s StringMap) Value() (driver.Value, error) {
	if s == nil {
		return "{}", nil
	}
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *StringMap) Scan(value interface{}) error {
	if value == nil {
		*s = StringMap{}
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal(value.([]byte), &m); err != nil {
		return err
	}
	*s = m
	return nil
}

type VolumeMount struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	Mode          string `json:"mode"`
}

type VolumeMountList []VolumeMount

func (v VolumeMountList) Value() (driver.Value, error) {
	if v == nil {
		return "[]", nil
	}
	b, err := json.Marshal(v)
	return string(b), err
}

func (v *VolumeMountList) Scan(value interface{}) error {
	if value == nil {
		*v = VolumeMountList{}
		return nil
	}
	var list []VolumeMount
	if err := json.Unmarshal(value.([]byte), &list); err != nil {
		return err
	}
	*v = list
	return nil
}

type StringList []string

func (s StringList) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = StringList{}
		return nil
	}
	var list []string
	if err := json.Unmarshal(value.([]byte), &list); err != nil {
		return err
	}
	*s = list
	return nil
}