@echo off

setlocal
cd "C:\Program Files\Mosquitto"

@echo on
mosquitto_sub -h localhost -p 1883 -t test -u admin -P tck66wXyQ4yvNPzeB5fI -i test


