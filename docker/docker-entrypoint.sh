#!/bin/sh

# Home Assistant add-on: read config from /data/options.json
if [ -f "/data/options.json" ]; then
    GROUPALARM_APIKEY=$(jq -r '.groupalarm_apikey' /data/options.json)
    GROUPALARM_ORGS=$(jq -r '.groupalarm_orgs' /data/options.json)
    MQTT_HOST=$(jq -r '.mqtt_host' /data/options.json)
    MQTT_PORT=$(jq -r '.mqtt_port' /data/options.json)
    MQTT_USER=$(jq -r '.mqtt_user // ""' /data/options.json)
    MQTT_PASSWORD=$(jq -r '.mqtt_password // ""' /data/options.json)
    MQTT_TOPIC=$(jq -r '.mqtt_topic' /data/options.json)

# Plain Docker: optionally load a .env file
elif [ -f "/app/.env" ]; then
    set -o allexport
    ENVFILE=$(cat /app/.env | tr -d '\r')
    echo $ENVFILE > /tmp/.env.tmp
    source /tmp/.env.tmp
    rm /tmp/.env.tmp
    set +o allexport
fi

cat <<EOF > /app/config.ini
[groupalarm]
api_key = ${GROUPALARM_APIKEY}
organisations = ${GROUPALARM_ORGS}

[mqtt]
host = ${MQTT_HOST}
port = ${MQTT_PORT:-1883}
user = ${MQTT_USER}
password = ${MQTT_PASSWORD}
client = groupalarm_app-$RANDOM
topic = ${MQTT_TOPIC:-"pager/groupalarm/{org}"}

EOF

exec /app/app
