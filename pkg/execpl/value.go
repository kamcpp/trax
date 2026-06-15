package execpl

import (
	"fmt"
	"strconv"
)

type Value struct {
	Key   string            `json:"key"`
	Type  string            `json:"type"`
	Value interface{}       `json:"value"`
	Extra map[string]string `json:"extra"`
}

func HasKey(values *map[string]*Value, key string) bool {
	if values == nil {
		return false
	}
	_, exists := (*values)[key]
	return exists
}

func InsertIntegerValue(values *map[string]*Value, key string, value int64) {
	if values == nil {
		return
	}
	(*values)[key] = &Value{
		Key:   key,
		Type:  string(ValueTypeEnum_Integer),
		Value: value,
		Extra: map[string]string{},
	}
}

func InsertFloat64Value(values *map[string]*Value, key string, value float64) {
	if values == nil {
		return
	}
	(*values)[key] = &Value{
		Key:   key,
		Type:  string(ValueTypeEnum_Float64),
		Value: value,
		Extra: map[string]string{},
	}
}

func InsertStringValue(values *map[string]*Value, key, value string) {
	if values == nil {
		return
	}
	(*values)[key] = &Value{
		Key:   key,
		Type:  string(ValueTypeEnum_String),
		Value: value,
		Extra: map[string]string{},
	}
}

func MustGetIntegerValue(values *map[string]*Value, key string) int64 {
	if values == nil {
		panic("values map is nil")
	}
	value, exists := (*values)[key]
	if !exists {
		panic(fmt.Sprintf("key is not found: '%s'", key))
	}
	if value.Type == string(ValueTypeEnum_String) {
		int64Val, err := strconv.ParseInt(value.Value.(string), 10, 64)
		if err != nil {
			panic(fmt.Sprintf("parsing int64 from string failed: [%s] %s",
				err.Error(), value.Value.(string)))
		}
		return int64Val
	} else if value.Type != string(ValueTypeEnum_Integer) {
		panic(fmt.Sprintf("value is not int64 for the key '%s': %v", key, value))
	}
	intVal, ok := value.Value.(int64)
	if !ok {
		floatVal, ok := value.Value.(float64)
		if !ok {
			panic(fmt.Sprintf("casting to int64 (via float64) failed for '%s' => %v", key, value))
		}
		intVal = int64(floatVal)
	}
	return intVal
}

func MustGetFloat64Value(values *map[string]*Value, key string) float64 {
	if values == nil {
		panic("values map is nil")
	}
	value, exists := (*values)[key]
	if !exists {
		panic(fmt.Sprintf("key is not found: '%s'", key))
	}
	if value.Type == string(ValueTypeEnum_String) {
		float64Val, err := strconv.ParseFloat(value.Value.(string), 64)
		if err != nil {
			panic(fmt.Sprintf("parsing float64 from string failed: [%s] %s",
				err.Error(), value.Value.(string)))
		}
		return float64Val
	} else if value.Type != string(ValueTypeEnum_Float64) {
		panic(fmt.Sprintf("value is not float64 for the key '%s': %v", key, value))
	}
	float64Val, ok := value.Value.(float64)
	if !ok {
		panic(fmt.Sprintf("casting to float64 failed for '%s' => %v", key, value))
	}
	return float64Val
}

func MustGetStringValue(values *map[string]*Value, key string) string {
	if values == nil {
		panic("values map is nil")
	}
	value, exists := (*values)[key]
	if !exists {
		panic(fmt.Sprintf("key is not found: '%s'", key))
	}
	str, ok := value.Value.(string)
	if !ok || value.Type != string(ValueTypeEnum_String) {
		panic(fmt.Sprintf("value is not string for the key '%s': %v", key, value))
	}
	return str
}

func GetStringValueWithDefaultValue(values *map[string]*Value, key, defaultValue string) string {
	if values == nil {
		panic("values map is nil")
	}
	value, exists := (*values)[key]
	if !exists {
		return defaultValue
	}
	str, ok := value.Value.(string)
	if !ok {
		panic(fmt.Sprintf("value is not string for the key '%s': %v", key, value))
	}
	return str
}
