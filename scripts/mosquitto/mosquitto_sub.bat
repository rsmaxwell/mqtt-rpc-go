@echo off

setlocal
cd %USERPROFILE%\git\github.com\rsmaxwell\diaries\server-go\scripts\mosquitto

@echo on
"C:\Program Files\Mosquitto\mosquitto_sub" -i subscriber -h localhost -t test -u richard -P secret