// Package api Generated by lanai-cli codegen. DO NOT EDIT
// Derived from openapi contract - components
package api

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type AdditionalPropertyTest struct {
	AttributeWithEmptyObjAP                  map[string]interface{}                                          `json:"attributeWithEmptyObjAP,omitempty"`
	AttributeWithFalseAP                     *AdditionalPropertyTestAttributeWithFalseAP                     `json:"attributeWithFalseAP,omitempty"`
	AttributeWithObjectPropertiesAndObjAP    *AdditionalPropertyTestAttributeWithObjectPropertiesAndObjAP    `json:"attributeWithObjectPropertiesAndObjAP,omitempty"`
	AttributeWithObjectPropertiesAndStringAP *AdditionalPropertyTestAttributeWithObjectPropertiesAndStringAP `json:"attributeWithObjectPropertiesAndStringAP,omitempty"`
	AttributeWithObjectPropertiesAndTrueAP   *AdditionalPropertyTestAttributeWithObjectPropertiesAndTrueAP   `json:"attributeWithObjectPropertiesAndTrueAP,omitempty"`
	AttributeWithTrueAP                      map[string]interface{}                                          `json:"attributeWithTrueAP,omitempty"`
}

type AdditionalPropertyTestAttributeWithFalseAP struct {
	Property string `json:"property"`
}

type AdditionalPropertyTestAttributeWithObjectPropertiesAndObjAP struct {
	Property string                 `json:"property"`
	Values   map[string]interface{} `json:"-"`
}

func (t *AdditionalPropertyTestAttributeWithObjectPropertiesAndObjAP) UnmarshalJSON(data []byte) (err error) {
	type ptrType *AdditionalPropertyTestAttributeWithObjectPropertiesAndObjAP
	if e := json.Unmarshal(data, ptrType(t)); e != nil {
		return e
	}
	if e := json.Unmarshal(data, &t.Values); e != nil {
		return e
	}
	return nil
}

func (t AdditionalPropertyTestAttributeWithObjectPropertiesAndObjAP) MarshalJSON() ([]byte, error) {
	type AdditionalPropertyTestAttributeWithObjectPropertiesAndObjAP_ AdditionalPropertyTestAttributeWithObjectPropertiesAndObjAP
	bytes, err := json.Marshal(AdditionalPropertyTestAttributeWithObjectPropertiesAndObjAP_(t))
	if err != nil {
		return nil, err
	}
	if t.Values == nil || len(t.Values) == 0 {
		return bytes, nil
	}
	extra, err := json.Marshal(t.Values)
	if err != nil {
		return nil, err
	}

	if string(bytes) == "{}" {
		return extra, nil
	}
	bytes[len(bytes)-1] = ','
	return append(bytes, extra[1:]...), nil
}

type AdditionalPropertyTestAttributeWithObjectPropertiesAndStringAP struct {
	Property string            `json:"property"`
	Values   map[string]string `json:"-"`
}

func (t *AdditionalPropertyTestAttributeWithObjectPropertiesAndStringAP) UnmarshalJSON(data []byte) (err error) {
	type ptrType *AdditionalPropertyTestAttributeWithObjectPropertiesAndStringAP
	if e := json.Unmarshal(data, ptrType(t)); e != nil {
		return e
	}
	if e := json.Unmarshal(data, &t.Values); e != nil {
		return e
	}
	return nil
}

func (t AdditionalPropertyTestAttributeWithObjectPropertiesAndStringAP) MarshalJSON() ([]byte, error) {
	type AdditionalPropertyTestAttributeWithObjectPropertiesAndStringAP_ AdditionalPropertyTestAttributeWithObjectPropertiesAndStringAP
	bytes, err := json.Marshal(AdditionalPropertyTestAttributeWithObjectPropertiesAndStringAP_(t))
	if err != nil {
		return nil, err
	}
	if t.Values == nil || len(t.Values) == 0 {
		return bytes, nil
	}
	extra, err := json.Marshal(t.Values)
	if err != nil {
		return nil, err
	}

	if string(bytes) == "{}" {
		return extra, nil
	}
	bytes[len(bytes)-1] = ','
	return append(bytes, extra[1:]...), nil
}

type AdditionalPropertyTestAttributeWithObjectPropertiesAndTrueAP struct {
	Property string                 `json:"property"`
	Values   map[string]interface{} `json:"-"`
}

func (t *AdditionalPropertyTestAttributeWithObjectPropertiesAndTrueAP) UnmarshalJSON(data []byte) (err error) {
	type ptrType *AdditionalPropertyTestAttributeWithObjectPropertiesAndTrueAP
	if e := json.Unmarshal(data, ptrType(t)); e != nil {
		return e
	}
	if e := json.Unmarshal(data, &t.Values); e != nil {
		return e
	}
	return nil
}

func (t AdditionalPropertyTestAttributeWithObjectPropertiesAndTrueAP) MarshalJSON() ([]byte, error) {
	type AdditionalPropertyTestAttributeWithObjectPropertiesAndTrueAP_ AdditionalPropertyTestAttributeWithObjectPropertiesAndTrueAP
	bytes, err := json.Marshal(AdditionalPropertyTestAttributeWithObjectPropertiesAndTrueAP_(t))
	if err != nil {
		return nil, err
	}
	if t.Values == nil || len(t.Values) == 0 {
		return bytes, nil
	}
	extra, err := json.Marshal(t.Values)
	if err != nil {
		return nil, err
	}

	if string(bytes) == "{}" {
		return extra, nil
	}
	bytes[len(bytes)-1] = ','
	return append(bytes, extra[1:]...), nil
}

type ApiPolicy struct {
	Unlimited bool `json:"unlimited,omitempty"`
}

type Device struct {
	CreatedOn                           *time.Time              `json:"createdOn,omitempty"`
	Id                                  *uuid.UUID              `json:"id,omitempty"`
	ModifiedOn                          *time.Time              `json:"modifiedOn,omitempty"`
	RegexWithBackslashes                *string                 `json:"regexWithBackslashes,omitempty" binding:"omitempty,regexA217A"`
	ServiceType                         string                  `json:"serviceType" binding:"omitempty,max=128"`
	Status                              DeviceStatus            `json:"status,omitempty"`
	StatusDetails                       map[string]DeviceStatus `json:"statusDetails,omitempty"`
	StringWithFormatOnlyInAnAllOfSchema *string                 `json:"stringWithFormatOnlyInAnAllOfSchema,omitempty" binding:"omitempty,regexEFF6B"`
	SubscriptionId                      *uuid.UUID              `json:"subscriptionId,omitempty"`
	UserId                              *uuid.UUID              `json:"userId,omitempty"`
}

type DeviceCreate struct {
	ServiceInstanceId uuid.UUID `json:"serviceInstanceId" binding:"required"`
	DeviceUpdate
}

type DeviceStatus struct {
	LastUpdated        time.Time `json:"lastUpdated" binding:"required"`
	LastUpdatedMessage string    `json:"lastUpdatedMessage" binding:"required,min=1,max=128"`
	Severity           string    `json:"severity" binding:"required,min=1,max=128"`
	Value              string    `json:"value" binding:"required,min=1,max=128"`
}

type DeviceUpdate struct {
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type GenericObject struct {
	Enabled        GenericObjectEnabled         `json:"enabled,omitempty"`
	Id             string                       `json:"id"`
	ValueWithAllOf *GenericObjectValueWithAllOf `json:"valueWithAllOf,omitempty"`
}

type GenericObjectEnabled struct {
	Inner string `json:"inner"`
}

type GenericObjectValueWithAllOf struct {
	ApiPolicy
}

type GenericResponse struct {
	ArrayOfObjects                  []GenericObject             `json:"arrayOfObjects"`
	ArrayOfRef                      []string                    `json:"arrayOfRef"`
	ArrayOfUUIDs                    []uuid.UUID                 `json:"arrayOfUUIDs" binding:"omitempty,dive,uuid"`
	CreatedOnDate                   string                      `json:"createdOnDate" binding:"required,date"`
	CreatedOnDateTime               *time.Time                  `json:"createdOnDateTime,omitempty"`
	DirectRef                       GenericObject               `json:"directRef,omitempty"`
	Email                           string                      `json:"email" binding:"omitempty,email"`
	EmptyObject                     map[string]interface{}      `json:"emptyObject,omitempty"`
	Integer32Value                  int32                       `json:"integer32Value" binding:"omitempty,max=5"`
	Integer64Value                  int64                       `json:"integer64Value"`
	IntegerValue                    int                         `json:"integerValue"`
	MyUuid                          *uuid.UUID                  `json:"myUuid,omitempty"`
	NumberArray                     []float64                   `json:"numberArray" binding:"omitempty,max=10"`
	NumberValue                     float64                     `json:"numberValue" binding:"omitempty,max=10"`
	ObjectValue                     *GenericResponseObjectValue `json:"objectValue" binding:"required"`
	StringValue                     *string                     `json:"stringValue" binding:"required,max=128"`
	StringWithEnum                  string                      `json:"stringWithEnum" binding:"omitempty,enumof=asc desc"`
	StringWithNilEnum               string                      `json:"stringWithNilEnum" binding:"omitempty,enumof=asc desc"`
	StringWithRegexDefinedInFormat  *string                     `json:"stringWithRegexDefinedInFormat,omitempty" binding:"omitempty,regex50C92"`
	StringWithRegexDefinedInPattern string                      `json:"stringWithRegexDefinedInPattern" binding:"required,regexA2EFE"`
	Values                          map[string]string           `json:"-"`
}

func (t *GenericResponse) UnmarshalJSON(data []byte) (err error) {
	type ptrType *GenericResponse
	if e := json.Unmarshal(data, ptrType(t)); e != nil {
		return e
	}
	if e := json.Unmarshal(data, &t.Values); e != nil {
		return e
	}
	return nil
}

func (t GenericResponse) MarshalJSON() ([]byte, error) {
	type GenericResponse_ GenericResponse
	bytes, err := json.Marshal(GenericResponse_(t))
	if err != nil {
		return nil, err
	}
	if t.Values == nil || len(t.Values) == 0 {
		return bytes, nil
	}
	extra, err := json.Marshal(t.Values)
	if err != nil {
		return nil, err
	}

	if string(bytes) == "{}" {
		return extra, nil
	}
	bytes[len(bytes)-1] = ','
	return append(bytes, extra[1:]...), nil
}

type GenericResponseObjectValue struct {
	ObjectNumber *float64 `json:"objectNumber" binding:"required"`
}

type GenericResponseWithAllOf struct {
	Id string `json:"id"`
	GenericResponse
}

type ObjectWithRefAndAdditionalProperties struct {
	Values map[string]string `json:"-"`
	GenericObject
}

func (t *ObjectWithRefAndAdditionalProperties) UnmarshalJSON(data []byte) (err error) {
	type ptrType *ObjectWithRefAndAdditionalProperties
	if e := json.Unmarshal(data, ptrType(t)); e != nil {
		return e
	}
	if e := json.Unmarshal(data, &t.Values); e != nil {
		return e
	}
	return nil
}

func (t ObjectWithRefAndAdditionalProperties) MarshalJSON() ([]byte, error) {
	type ObjectWithRefAndAdditionalProperties_ ObjectWithRefAndAdditionalProperties
	bytes, err := json.Marshal(ObjectWithRefAndAdditionalProperties_(t))
	if err != nil {
		return nil, err
	}
	if t.Values == nil || len(t.Values) == 0 {
		return bytes, nil
	}
	extra, err := json.Marshal(t.Values)
	if err != nil {
		return nil, err
	}

	if string(bytes) == "{}" {
		return extra, nil
	}
	bytes[len(bytes)-1] = ','
	return append(bytes, extra[1:]...), nil
}

type RequestBodyWithAllOf struct {
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Managed    bool                   `json:"managed,omitempty"`
}

type TestRequest struct {
	Uuid *uuid.UUID `json:"uuid,omitempty"`
}
