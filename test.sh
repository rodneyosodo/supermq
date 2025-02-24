#!/bin/bash

USER_TOKEN="eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mzg3OTI5NDMsImlhdCI6MTczODc4OTM0MywiaXNzIjoic3VwZXJtcS5hdXRoIiwidHlwZSI6MCwidXNlciI6IjdjYmVlZTNkLTU5NGYtNGMyMS04MTI4LTlmYWUyOTdlZTZjYyJ9.uKh2FnU3-I_1jdgqgVqW8PJGFrsFcZ2WQKTerKMPxICnqcFQTfLqGYM-2yC2yWBHI2BNich8p14S0DBp7gRP_A"

DOMAIN_ID="76e4c760-0d78-45ca-a93c-c83db14e809e"

THING_ID="3b0c3d98-a2dd-439b-891b-03aceec88306"
supermq-cli clients get $THING_ID $DOMAIN_ID $USER_TOKEN

export THING_ID="3b0c3d98-a2dd-439b-891b-03aceec88306"
export THING_KEY="070067af-b32b-4f4b-9f39-55b609fca2ca"
export CHANNEL_ID="17126966-3e9c-4d7f-9ec5-c9204701b859"

coap-cli get channels/$CHANNEL_ID/messages -a $THING_KEY -o

mosquitto_sub -u $THING_ID -P $THING_KEY -t channels/$CHANNEL_ID/messages -I supermq -h localhost

ws://localhost:8186/channels/17126966-3e9c-4d7f-9ec5-c9204701b859/messages?authorization=070067af-b32b-4f4b-9f39-55b609fca2ca

for i in {1..10}; do
  TIMESTAMP=$(date +%s)
  mosquitto_pub -u $THING_ID -P $THING_KEY -t channels/$CHANNEL_ID/messages -I supermq -h localhost -m "[{\"bn\":\"mqtt-device:\",\"bt\":$TIMESTAMP, \"bu\":\"A\",\"bver\":5, \"n\":\"voltage\",\"u\":\"V\",\"v\":120.1, \"i\":$i}, {\"n\":\"current\",\"t\":-5,\"v\":1.2, \"i\":$i}, {\"n\":\"current\",\"t\":-4,\"v\":1.3, \"i\":$i}]"
  sleep 1
done

coap-cli post channels/$CHANNEL_ID/messages -a $THING_KEY -d "[{\"bn\":\"coap-device:\",\"bt\":$TIMESTAMP, \"bu\":\"A\",\"bver\":5, \"n\":\"voltage\",\"u\":\"V\",\"v\":120.1, \"i\":$i}, {\"n\":\"current\",\"t\":-5,\"v\":1.2, \"i\":$i}, {\"n\":\"current\",\"t\":-4,\"v\":1.3, \"i\":$i}]"

mosquitto_pub -u $THING_ID -P $THING_KEY -t channels/$CHANNEL_ID/messages -I supermq -h localhost -m "[{\"bn\":\"mqtt-device:\",\"bt\":$TIMESTAMP, \"bu\":\"A\",\"bver\":5, \"n\":\"voltage\",\"u\":\"V\",\"v\":120.1, \"i\":$i}, {\"n\":\"current\",\"t\":-5,\"v\":1.2, \"i\":$i}, {\"n\":\"current\",\"t\":-4,\"v\":1.3, \"i\":$i}]"

curl -i -X POST -H "Content-Type: application/senml+json" -H "Authorization: Client $THING_KEY" http://localhost/http/channels/$CHANNEL_ID/messages -d "[{\"bn\":\"http-device:\",\"bt\":$TIMESTAMP, \"bu\":\"A\",\"bver\":5, \"n\":\"voltage\",\"u\":\"V\",\"v\":120.1, \"i\":$i}, {\"n\":\"current\",\"t\":-5,\"v\":1.2, \"i\":$i}, {\"n\":\"current\",\"t\":-4,\"v\":1.3, \"i\":$i}]"
