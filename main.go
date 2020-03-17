package main

import (
	"context"
	"flag"
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
	exit   context.CancelFunc
}

func (p *program) Start(s service.Service) error {
	ctx, stop := context.WithCancel(context.Background())
	p.ctx = ctx
	p.exit = stop
	go p.run(s)
	return nil
}

func (p *program) run(s service.Service) {
	logger.Info("Starting")
	defer func() {
		if service.Interactive() {
			p.Stop(s)
		} else {
			s.Stop()
		}
	}()

	influxClient, err := p.startInfluxClient()
	if err != nil {
		// If we got an error here, there is probably something with the connection to the InfluxDB.
		// TODO: Retry later.
		logger.Error(err)
		return
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
	p.exit()
	logger.Info("Stopping")
	if service.Interactive() {
		os.Exit(0)
	}
	return nil
}

func main() {
	options := make(service.KeyValue)
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 SIGKILL"
	svcConfig := &service.Config{
		Name:        "FroniusDataDump",
		DisplayName: "Fronius Data Dump",
		Description: "Read data from a Fronius product and store it in InfluxDb.",
		Dependencies: []string{
			"Requires=network.target",
			"After=network-online.target syslog.target"},
		Option: options,
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

	showHelp, svcFlag := prg.readConfig()
	if showHelp {
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}
		log.Println("Get information from Fronius solar api and insert it into an InfluxDb.")
		log.Println()
		log.Printf("Could not read config file in current working directory '%s'.\n", path)
		log.Printf("Please creata a file named '%s' there, looking like this:\n", configFileName)
		log.Println("{")
		log.Println("	\"Source\": {")
		log.Println("		\"APIUri\": \"http://localhost/solar_api/v1/\",")
		log.Println("		\"SleepInSeconds\": 10")
		log.Println("	},")
		log.Println("	\"Sink\": {")
		log.Println("		\"APIUri\": \"http://192.168.2.80:8086/\",")
		log.Println("		\"Database\": \"mydb\"")
		log.Println("	},")
		log.Println("	\"VerboseLogging\": false")
		log.Println("}")
		log.Println()
		log.Printf("Valid actions: %q\n", service.ControlAction)
		log.Println()
		return
	}

	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}

	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}

// readConfig setup and read configurations from file or commandline arguments.
func (p *program) readConfig() (showHelp bool, svcFlag *string) {
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

	svcFlag = flag.String("service", "", "Control the system service.")
	flag.Parse()

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
