#!/bin/sh
set -e
sleep 5 # Workaround to wait untill InfluxDb will start
/root/statusok --config /config/config.json
