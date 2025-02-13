#!/bin/bash

USER_TOKEN="eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzkzOTM0NTUsImlhdCI6MTczOTM4OTg1NSwiaXNzIjoic3VwZXJtcS5hdXRoIiwidHlwZSI6MCwidXNlciI6IjFlYTVhMzI2LTUyNDEtNGQzYy05Y2ZlLTQ5Y2E5OWYxYWI3MSJ9.gNoAnJaKMjJS75md27o1m-wesVdijBMjLtTJOZxrXr8NpIevm0R53CmVuqENYUIIQyvNppE8kNnSDz8SpI-a6g"


DOMAIN_ID="e16c36d2-0cc1-45cb-9f1e-0398a7d1ddad"

THING_ID="09e87117-120c-4f18-aede-fbeefb42b9f4"
supermq-cli clients get $THING_ID $DOMAIN_ID $USER_TOKEN

THING_ID_2="53d8e60b-84ee-495e-b9da-477e81fa9177"
supermq-cli clients get $THING_ID_2 $DOMAIN_ID $USER_TOKEN

THING_KEY_2="43cf5966-8d6d-498d-ad1b-f840ac23e2fc"

export THING_ID="09e87117-120c-4f18-aede-fbeefb42b9f4"
export THING_KEY="cd8934fd-1c43-44f3-9231-a4bee5b90922"
export CHANNEL_ID="fa9b8aaf-5038-4f50-8586-17964a6bc0d0"

coap-cli get channels/$CHANNEL_ID/messages -a $THING_KEY -o

mosquitto_sub -u $THING_ID -P $THING_KEY -t channels/$CHANNEL_ID/messages -I supermq -h localhost

ws://localhost:8186/channels/fa9b8aaf-5038-4f50-8586-17964a6bc0d0/messages?authorization=cd8934fd-1c43-44f3-9231-a4bee5b90922
ws://localhost:8186/channels/fa9b8aaf-5038-4f50-8586-17964a6bc0d0/messages?authorization=43cf5966-8d6d-498d-ad1b-f840ac23e2fc

for i in {1..1000000}; do
  TIMESTAMP=$(date +%s)
  coap-cli post channels/$CHANNEL_ID/messages -a $THING_KEY -d "[{\"bn\":\"coap-device:\",\"bt\":$TIMESTAMP, \"bu\":\"A\",\"bver\":5, \"n\":\"voltage\",\"u\":\"V\",\"v\":120.1, \"i\":$i}, {\"n\":\"current\",\"t\":-5,\"v\":1.2, \"i\":$i}, {\"n\":\"current\",\"t\":-4,\"v\":1.3, \"i\":$i}]"
  mosquitto_pub -u $THING_ID -P $THING_KEY -t channels/$CHANNEL_ID/messages -I supermq -h localhost -m "[{\"bn\":\"mqtt-device:\",\"bt\":$TIMESTAMP, \"bu\":\"A\",\"bver\":5, \"n\":\"voltage\",\"u\":\"V\",\"v\":120.1, \"i\":$i}, {\"n\":\"current\",\"t\":-5,\"v\":1.2, \"i\":$i}, {\"n\":\"current\",\"t\":-4,\"v\":1.3, \"i\":$i}]"
done

coap-cli post channels/$CHANNEL_ID/messages -a $THING_KEY -d "[{\"bn\":\"coap-device:\",\"bt\":$TIMESTAMP, \"bu\":\"A\",\"bver\":5, \"n\":\"voltage\",\"u\":\"V\",\"v\":120.1, \"i\":$i}, {\"n\":\"current\",\"t\":-5,\"v\":1.2, \"i\":$i}, {\"n\":\"current\",\"t\":-4,\"v\":1.3, \"i\":$i}]"

mosquitto_pub -u $THING_ID -P $THING_KEY -t channels/$CHANNEL_ID/messages -I supermq -h localhost -m "[{\"bn\":\"mqtt-device:\",\"bt\":$TIMESTAMP, \"bu\":\"A\",\"bver\":5, \"n\":\"voltage\",\"u\":\"V\",\"v\":120.1, \"i\":$i}, {\"n\":\"current\",\"t\":-5,\"v\":1.2, \"i\":$i}, {\"n\":\"current\",\"t\":-4,\"v\":1.3, \"i\":$i}]"

curl -i -X POST -H "Content-Type: application/senml+json" -H "Authorization: Client $THING_KEY" http://localhost/http/channels/$CHANNEL_ID/messages -d "[{\"bn\":\"http-device:\",\"bt\":$TIMESTAMP, \"bu\":\"A\",\"bver\":5, \"n\":\"voltage\",\"u\":\"V\",\"v\":120.1, \"i\":$i}, {\"n\":\"current\",\"t\":-5,\"v\":1.2, \"i\":$i}, {\"n\":\"current\",\"t\":-4,\"v\":1.3, \"i\":$i}]"

export THING_ID="supermq"
export THING_KEY="supermq"
export CHANNEL_ID="a8110477-8cce-43c9-a1c4-5fd3230ce038"


curl -i -X POST -H "Content-Type: application/senml+json" -H "Authorization: Client $THING_KEY" http://localhost/http/channels/$CHANNEL_ID/messages -d "[{\"bn\":\"http-device:\",\"bt\":$TIMESTAMP, \"bu\":\"A\",\"bver\":5, \"n\":\"voltage\",\"u\":\"V\",\"v\":120.1, \"i\":$i}, {\"n\":\"current\",\"t\":-5,\"v\":1.2, \"i\":$i}, {\"n\":\"current\",\"t\":-4,\"v\":1.3, \"i\":$i}, {\"n\":\"temperature\",\"t\":-3,\"u\":\"C\",\"v\":25.3, \"i\":$i}, {\"n\":\"humidity\",\"t\":-2,\"u\":\"%RH\",\"v\":60.5, \"i\":$i}, {\"n\":\"power\",\"t\":-1,\"u\":\"W\",\"v\":150.0, \"i\":$i}, {\"n\":\"energy\",\"t\":0,\"u\":\"Wh\",\"v\":5000.2, \"i\":$i}, {\"n\":\"frequency\",\"t\":1,\"u\":\"Hz\",\"v\":50.0, \"i\":$i}, {\"n\":\"pressure\",\"t\":2,\"u\":\"Pa\",\"v\":101325, \"i\":$i}]"
