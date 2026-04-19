# GroupAlarm MQTT Bridge — Home Assistant Add-on

Polls the [GroupAlarm](https://app.groupalarm.com) API every 5 seconds for open alarms and publishes new alarms as MQTT messages.

## Installation

1. In Home Assistant, navigate to **Settings → Add-ons → Add-on Store**.
2. Click the three-dot menu (⋮) and choose **Repositories**.
3. Add the URL of this repository and click **Add**.
4. Find **GroupAlarm MQTT Bridge** in the store and click **Install**.

## Configuration

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `groupalarm_apikey` | Yes | — | Your GroupAlarm Personal Access Token |
| `groupalarm_orgs` | Yes | — | Comma-separated organisation IDs to poll (e.g. `19184,19185`) |
| `mqtt_host` | Yes | `core-mosquitto` | MQTT broker hostname. Use `core-mosquitto` for the built-in HA broker |
| `mqtt_port` | Yes | `1883` | MQTT broker port |
| `mqtt_user` | No | — | MQTT username |
| `mqtt_password` | No | — | MQTT password |
| `mqtt_topic` | Yes | `pager/groupalarm/{org}` | MQTT topic template. `{org}` is replaced with the organisation ID |

## MQTT Payload

Each alarm is published as the raw alarm message string to the configured topic with QoS 1.

**Example topic:** `pager/groupalarm/19184`
**Example payload:** `Alarm: Fire in building A`

## Notes

- Alarm IDs are tracked in memory only. Restarting the add-on will cause already-seen alarms from open events to be re-published.
- The add-on connects to the built-in Mosquitto broker by default (`core-mosquitto`). Install the **Mosquitto broker** add-on if you haven't already.
