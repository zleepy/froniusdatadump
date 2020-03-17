# froniusdatadump
FDD is a simple service that reads from a Fronius web API and delivers the data to an InfluxDB.

Tjenare,
Det här programmet går att köra fristående eller som en windows service.Det där med Windows service har jag inte testat än så det garanterar jag inte att det fungerar.
Men du kan väl testa det i alla fall genom att köra det från kommandoprompten eller som schemalagt jobb på servern. Kontrollera sökvägarna i config.json först. Den filen måste ligga i den katalog som är aktiv när programmet köras. Det är oftast samma som exe-filen ligger i om man inte specifikt ändrar det.
Det loggar i stort sett alla värden som hittas per site och per inverter för GetPowerFlowRealtimeData, samt per instans den hittar under Data för GetMeterRealtimeData.




## InfluxDB things
``` bahsrc
$ influx -precision rfc3339
> show databases
…

> create database mydb

> use mydb
Using database mydb

> insert cpu,host=RPi value=0.64

> select "host","value" from "cpu"
name: cpu
---------
time                            host    value

2020-03-03T21:00:17.517428683Z  RPi     0.64
```

# Fronius phase
http://192.168.2.69/solar_api/v1/GetMeterRealtimeData.cgi?Scope=System

# Fronius power
http://192.168.2.69/solar_api/v1/GetPowerFlowRealtimeData.fcgi




# Avinstallerat gamla versionen av GO

## Användarvariabler
PATH -`C:\Users\Maria\Documents\Anders\Projekt\go\bin`
GOOS -`linux`
GOPATH -+`%USERPROFILE%\Documents\Anders\Projekt\go`

## Systemvariabler
GOROOT -`C:\Go\`
GOPATH -`C:\Users\Maria\Documents\Anders\Projekt\go`


# WLS installation av senaste versionen av GO
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