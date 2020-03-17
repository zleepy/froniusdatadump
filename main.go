package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/kardianos/service"
)

var logger service.Logger

type program struct {
	config configurations
	ctx    context.Context
	stop   context.CancelFunc
}

func (p *program) Start(s service.Service) error {
	ctx, stop := context.WithCancel(context.Background())
	p.ctx = ctx
	p.stop = stop
	go p.run()
	return nil
}

func (p *program) run() {
	influxClient, err := p.startInfluxClient()
	if err != nil {
		// If we got an error here, there is probably something with the connection to the InfluxDB.
		// TODO: Retry later.
		panic(logger.Error(err))
	}
	froniusClient := NewFronius(p.config.Source.APIUri)

	// Loop forever
	for {
		if err = p.extractAndSave(froniusClient, influxClient); err != nil {
			logger.Error(err)
		}

		select {
		case <-p.ctx.Done():
			return
		case <-time.After(time.Duration(p.config.Source.SleepInSec) * time.Second):
			// Loop
		}
	}
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	fmt.Println("Get information from Fronius solar api and insert it into an InfluxDb.")
	fmt.Println()

	svcConfig := &service.Config{
		Name:        "FroniusDataDump",
		DisplayName: "Fronius Data Dump",
		Description: "Hämtar ut data från en Fronius-produkt och lagrar i InfluxDb.",
	}

	prg := &program{}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if prg.readConfig() {
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("Could not read config file in current working directory '%s'.\n", path)
		fmt.Printf("Please creata a file named '%s' there, looking like this:\n", configFileName)
		fmt.Println("{")
		fmt.Println("	\"Source\": {")
		fmt.Println("		\"APIUri\": \"http://localhost/solar_api/v1/\",")
		fmt.Println("		\"SleepInSeconds\": 10")
		fmt.Println("	},")
		fmt.Println("	\"Sink\": {")
		fmt.Println("		\"APIUri\": \"http://192.168.2.80:8086/\",")
		fmt.Println("		\"Database\": \"mydb\"")
		fmt.Println("	},")
		fmt.Println("	\"VerboseLogging\": false")
		fmt.Println("}")
		return
	}

	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}

// readConfig setup and read configurations from file or commandline arguments.
func (p *program) readConfig() (showHelp bool) {
	// Default values
	p.config = configurations{
		Source: source{
			APIUri:     "http://localhost/solar_api/v1/",
			SleepInSec: 10,
		},
		Sink: sink{
			APIUri:   "http://192.168.2.80:8086/",
			Database: "mydb",
		},
		VerboseLogging: false,
	}
	err := readConfigFile(&p.config)
	if err != nil {
		showHelp = true
	}
	return

	// flag.String("fronius-addr", "f", "", "Address to Fronius api device")
	// flag.IntP("fronius-port", "p", 80, "Fronius port")
	// 	flag.IntP("sleep-in-seconds", "s", 10, "Seconds to sleep between collecting values")
	// 	flag.StringP("influx-addr", "i", "localhost", "Address to InfluxDB server")
	// 	flag.IntP("influx-port", "P", 8086, "InfluxDB port")
	// 	flag.StringP("influx-database", "d", "", "InfluxDB database name to store values in")
	// 	showHelp := flag.BoolP("help", "h", false, "Show help message")
	// 	flag.Parse()

}

//TODO: Extract to own file and interface?
// startInfluxClient setup and tests an InfluxDB client.
func (p *program) startInfluxClient() (client.Client, error) {
	influxClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:      p.config.Sink.APIUri,
		UserAgent: "FDD",
	})
	if err != nil {
		logger.Errorf("Error creating InfluxDB Client: %s", err.Error())
		return nil, err
	}
	defer influxClient.Close()

	duration, version, err := influxClient.Ping(0)
	if err != nil {
		logger.Error("Error creating InfluxDB Client: ", err.Error())
		return nil, err
	}

	logger.Infof("InfluxDB version %s, ping took %s, ", version, duration)
	return influxClient, nil
}

//TODO: Extract to own file and interface?
// extractAndSave hämtar ut information från en Fronius-klient och skickar in det till en InfluxDb.
func (p *program) extractAndSave(froniusClient *Fronius, influxClient client.Client) error {
	bps, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  p.config.Sink.Database,
		Precision: "s",
	})
	if err != nil {
		// If we got an error here something is seriously wrong.
		panic(logger.Error(err))
	}

	points, err := froniusClient.Extract()
	if err != nil {
		return err
	}

	if len(points) > 0 {
		bps.AddPoints(points)

		if err = influxClient.Write(bps); err != nil {
			return err
		}
		if p.config.VerboseLogging {
			logger.Info("Sent points to InfluxDb:", bps)
		}
	}
	return nil
}
