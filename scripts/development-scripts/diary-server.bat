@echo off

setlocal
cd %~dp0

echo on
diary-server.exe -username %MQTT_USERNAME% -password %MQTT_PASSWORD%
