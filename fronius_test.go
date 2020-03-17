package main

import (
	"io/ioutil"
	"testing"
)

func TestGetPointsFromGetMeterRealtimeDataJSON(t *testing.T) {
	var data, err = ioutil.ReadFile("testfiles/GetMeterRealtimeData.json")
	if err != nil {
		t.Fatal(err)
	}

	result, err := getPointsFromGetMeterRealtimeDataJSON(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 {
		t.Fatalf("Should be 1 was %d", len(result))
	}

	tags := result[0].Tags()
	if tags[FunctionKey] != "GetMeterRealtimeData" {
		t.Errorf("tag %s != \"GetMeterRealtimeData\" (%s)", FunctionKey, tags[FunctionKey])
	}
	if tags[DeviceIDKey] != "0" {
		t.Errorf("tag %s != \"0\" (%s)", DeviceIDKey, tags[DeviceIDKey])
	}

	fields, _ := result[0].Fields()
	const expectedLength = 35
	if len(fields) != expectedLength {
		t.Errorf("Should be %d was %d", expectedLength, len(fields))
	}
}

func TestGetPointsFromGetPowerFlowRealtimeDataJSON(t *testing.T) {
	var data, err = ioutil.ReadFile("testfiles/GetPowerFlowRealtimeData.json")
	if err != nil {
		t.Fatal(err)
	}

	result, err := getPointsFromGetPowerFlowRealtimeDataJSON(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 2 {
		t.Fatalf("Should be 2 was %d", len(result))
	}

	// Test Site values
	tags := result[0].Tags()
	if tags[FunctionKey] != "GetPowerFlowRealtimeData" {
		t.Errorf("Site tag %s != \"GetPowerFlowRealtimeData\" (%s)", FunctionKey, tags[FunctionKey])
	}
	if tags[InverterIDKey] != "Site" {
		t.Errorf("Site tag %s != \"Site\" (%s)", InverterIDKey, tags[InverterIDKey])
	}

	fields, _ := result[0].Fields()
	const expectedSiteLength = 8 // There are 11 fields but those with null is not added
	if len(fields) != expectedSiteLength {
		t.Errorf("Site should have %d values, had %d", expectedSiteLength, len(fields))
	}

	// Test Inverter 1 values
	tags = result[1].Tags()
	if tags[FunctionKey] != "GetPowerFlowRealtimeData" {
		t.Errorf("Inverter 1 tag %s != \"GetPowerFlowRealtimeData\" (%s)", FunctionKey, tags[FunctionKey])
	}
	if tags[InverterIDKey] != "1" {
		t.Errorf("Inverter 1 tag %s != \"1\" (%s)", InverterIDKey, tags[InverterIDKey])
	}

	fields, _ = result[1].Fields()
	const expectedInverterLength = 5
	if len(fields) != expectedInverterLength {
		t.Errorf("Inverter 1 should have %d values, had %d", expectedInverterLength, len(fields))
	}
}
