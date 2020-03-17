![Build](https://github.com/zleepy/froniusdatadump/workflows/Build/badge.svg)

# froniusdatadump
FDD is a simple service that reads from a Fronius web API and delivers the data to an InfluxDB.

## Instructions

This application can be run standalone from Linux or Windows command line. It should also be able to run as a service, but that functionality is not tested yet.

If not possible to install as a service you could schedule i to run using `cron` on linux and `Scheduler` on Windows. 

Before starting, verify settings in `config.json`. That file must lie in the active working directory, that is usualy in the same directory as the binary unless explicitly changed.

## What is logged
Most values for GetPowerFlowRealtimeData, per site and inverter.
Most values for GetMeterRealtimeData, per instance found under Data.


## InfluxDB things
How you could see logged values.

``` bahsrc
$ influx -precision rfc3339
> show databases
…

> create database mydb

> use mydb
Using database mydb

> insert cpu,host=RPi value=0.64

> select * from "fronius"

```

# My own notes 
## Fronius phase
http://froniusdevice/solar_api/v1/GetMeterRealtimeData.cgi?Scope=System

## Fronius power
http://froniusdevice/solar_api/v1/GetPowerFlowRealtimeData.fcgi

## WLS installation av senaste versionen av GO
Install Binary Distribution
Let’s say we wanted to get the latest version, we would have to install the official binary distribution.

We first go to https://golang.org/dl/ and select a compatible release according to our OS

go1.11.5.linux-amd64.tar.gz
Let us open our terminal download our file and extract it in our current directory.
```
wget https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz
$ sudo tar -xvf go1.11.5.linux-amd64.tar.gz
```
tar : stores/extracts files from the archive
x : extracts the file from target
v : verbosly list files process
f : use archive file
Now we have a folder /go on our current directory. Let us move it to /usr/local/

```
sudo mv go /usr/local
```
Set Enviroments
Our last step is to add our global variables on our .bahsrc or .profile file.

```
sudo nano ~/.bashrc
```
scroll down and add these to your .bashrc profile

```
# Go Global variables
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```
To save it Ctrl + o, and to exit nano Ctrl + x

update current session

```
source ~/.bashrc
```
Check
```
go version
// go version go1.11.5 linux/amd64
```
