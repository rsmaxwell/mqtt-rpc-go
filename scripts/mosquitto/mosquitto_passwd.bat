@echo off

setlocal
cd %USERPROFILE%\Mosquitto

@echo on
copy pwfile.source.txt pwfile.txt
"C:\Program Files\Mosquitto\mosquitto_passwd" -U pwfile.txt


