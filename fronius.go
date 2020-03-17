package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
)

// FunctionKey is the InfluxDB key with the name of the function used to retrieve the information.
const FunctionKey = "function"

// DeviceIDKey is the InfluxDB key for a device id.
const DeviceIDKey = "deviceId"

// InverterIDKey is the InfluxDB key för an inverter id.
const InverterIDKey = "inverterId"

// Fronius retrieves data from Fronius devices.
type Fronius struct {
	apiURI string
}

// NewFronius creates an object to use for retrieval of data via Fronius Solar API.
func NewFronius(apiURI string) *Fronius {
	return &Fronius{
		apiURI: apiURI,
	}
}

// Extract returns a batch of value points from a Fronius device.
func (f *Fronius) Extract() ([]*client.Point, error) {
	endpoints := map[string]func(data []byte) ([]*client.Point, error){
		"GetMeterRealtimeData.cgi?Scope=System": getPointsFromGetMeterRealtimeDataJSON,
		"GetPowerFlowRealtimeData.fcgi":         getPointsFromGetPowerFlowRealtimeDataJSON,
	}

	var result []*client.Point

	for function, extractFunc := range endpoints {
		points, err := f.getPointsFromAPI(function, extractFunc)
		if err != nil {
			return result, err
		}
		result = append(result, points...)
	}

	return result, nil
}

func (f *Fronius) getPointsFromAPI(function string, extractor func(data []byte) ([]*client.Point, error)) ([]*client.Point, error) {
	/*
		if strings.HasPrefix(function, "GetMeter") {
			body, err := ioutil.ReadFile("testfiles/GetMeterRealtimeData.json")
			if err != nil {
				return nil, err
			}
			return extractor(body)
		} else if strings.HasPrefix(function, "GetPower") {
			body, err := ioutil.ReadFile("testfiles/GetPowerFlowRealtimeData.json")
			if err != nil {
				return nil, err
			}
			return extractor(body)
		}
	*/
	//TODO: Use http.Client instead?
	uri := f.apiURI
	if !strings.HasSuffix(f.apiURI, "/") {
		uri = uri + "/"
	}
	resp, err := http.Get(fmt.Sprintf("%s%s", uri, function))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return extractor(body)
}

type getMeterRealtimeDataJSONBase struct {
	Head getJSONHead
	Body getMeterRealtimeDataJSNOBody
}

type getJSONHead struct {
	Timestamp time.Time
}

type getMeterRealtimeDataJSNOBody struct {
	Data map[string]map[string]interface{}
}

func getPointsFromGetMeterRealtimeDataJSON(data []byte) ([]*client.Point, error) {
	var o getMeterRealtimeDataJSONBase

	err := json.Unmarshal(data, &o)
	if err != nil {
		return nil, err
	}

	var result []*client.Point

	// Loop over all devices under Body/Data
	for k, v := range o.Body.Data {
		tags := map[string]string{
			FunctionKey: "GetMeterRealtimeData",
			DeviceIDKey: k,
		}

		// Add all fields found under a device (except TimeStamp and Details)
		fields := make(map[string]interface{})
		for fieldName, fieldValue := range v {
			if fieldValue != nil && fieldName != "TimeStamp" && fieldName != "Details" {
				fields[fieldName] = fieldValue
			}
		}

		p, err := client.NewPoint("fronius", tags, fields, o.Head.Timestamp)
		if err != nil {
			return nil, err
		}

		result = append(result, p)
	}

	return result, nil
}

type getPowerFlowRealtimeDataJSONBase struct {
	Head getJSONHead
	Body getPowerFlowRealtimeDataJSNOBody
}

type getPowerFlowRealtimeDataJSNOBody struct {
	Data getPowerFlowRealtimeDataJSNOData
}

type getPowerFlowRealtimeDataJSNOData struct {
	Site      map[string]interface{}
	Inverters map[string]map[string]interface{}
}

func getPointsFromGetPowerFlowRealtimeDataJSON(data []byte) ([]*client.Point, error) {
	var o getPowerFlowRealtimeDataJSONBase

	err := json.Unmarshal(data, &o)
	if err != nil {
		return nil, err
	}

	tags := map[string]string{
		FunctionKey:   "GetPowerFlowRealtimeData",
		InverterIDKey: "Site",
	}

	// Add all fields found under Site
	fields := make(map[string]interface{})
	for fieldName, fieldValue := range o.Body.Data.Site {
		if fieldValue != nil {
			fields[fieldName] = fieldValue
		}
	}

	p, err := client.NewPoint("fronius", tags, fields, o.Head.Timestamp)
	if err != nil {
		return nil, err
	}

	result := []*client.Point{p}

	// Loop over all devices under Body/Data/Inverters
	for k, v := range o.Body.Data.Inverters {
		tags = map[string]string{
			FunctionKey:   "GetPowerFlowRealtimeData",
			InverterIDKey: k,
		}

		// Add all fields found under an inverter
		fields = make(map[string]interface{})
		for fieldName, fieldValue := range v {
			if fieldValue != nil {
				fields[fieldName] = fieldValue
			}
		}

		p, err := client.NewPoint("fronius", tags, fields, o.Head.Timestamp)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}

	return result, nil
}
