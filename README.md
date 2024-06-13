"# diary-server-go" 

# Overview

The *diaries* application is a shared web app which automates the transcribing and viewing of images of historic documents. Users can view an image of a page and transcribe it into *html* so the document clearly viewed. Someone viewing a page will see the text change dynamically as someone else edits the same page.  This kind of webapp needs both a request/resoonse and a publish/subscribe messaging model, which is provided my MQTT v5. 

*MQTT* requires a broker process to be running to which apps connect. Although all the apps connecting to the *MQTT* broker are its clients, one of these apps is server for the webapp, managing a database of document images and html fragments, with which the webapp clients interact via the *MQTT* boker topic tree.

The *MQTT* broker used in this app is **Mosquitto** for the broker. Connections to the *MQTT* broker, which will be exposed to the internet, will be securing using password based authentication controlled with an *acl-file*, which is part of the **Mosquitto** configuration.

The webapp client will consist of the following separate parts:
    Calendar - to select a date 
    Viewer - to view an image of the diary page
    Editor - to edit the transcription of a diary page

The webapp client will require the user to sign in, then the user may:
    select an image of a diary page, then create/edit a transcription of a day
    select a date from the calendar, view the corresponding image, and create/edit a transcription of a day 

There will be separate user roles for editers and viewers, controled by the *acl-file*, so a viewer will only be able to view transcripts  

The webapp client will be implements in Angular, and the server in some appropiate programming environment, running in ciontainers in a docker compose environment.

But for now, this project is a “proof of concept” of a project written in GO which comprise of a server and a client implementing both a request/reply & pub/sub message exchange.  Other equivilent projects will be created implementing the same but written in different programming environments (e.g. Spring) such that each type of server works with each type of client. 

# Setup

The development environment will be Windows but the app will be deployed as a set of docker images in a docker compose environment on a linux machine.  

These instruction describe how to setup the components on a Windows machine.

## Mosquitto - MQTT Broker
 
    Install Mosquitto from   https://mosquitto.org/download/

    Make sure the scripts in the %PROJECT_DIR%\scripts\mosquitto directory correctly refer to where mosqitto is installed

    Start a new terminal (i.e. cmd.exe )
    Run %PROJECT_DIR%\scripts\mosquitto\mosquitto

## Build

    Install GO from https://go.dev/doc/install

    Set GOPATH to your installation, e.g.    set GOPATH=%USERPROFILE%/go

    Make sure that your PATH includes:     %GOPATH%\bin

    Run %PROJECT_DIR%\scripts\build, (which will build and install the server and client to:  %GOPATH%\bin )


## Server

    Start a new terminal (i.e. cmd.exe )
    Run %PROJECT_DIR%\scripts\development-scripts\diary-server

## Client

    Start a new terminal (i.e. cmd.exe )
    Run %PROJECT_DIR%\scripts\development-scripts\diary-server
