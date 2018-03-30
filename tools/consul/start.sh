#!/usr/bin/env bash
nohup ./consul  agent -config-dir ./conf -pid-file=./consul-server.pid &
