@echo off

setlocal
cd %USERPROFILE%\git\github.com\rsmaxwell\diaries\server-go\scripts\mosquitto

@echo on
"C:\Program Files\Mosquitto\mosquitto_pub" -i publisher -h localhost -t test -m "hello world"  -u richard -P secret


