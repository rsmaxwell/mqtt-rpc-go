@echo off

setlocal
cd %~dp0

echo on
diary-client.exe -username %MQTT_USERNAME% -password %MQTT_PASSWORD%
