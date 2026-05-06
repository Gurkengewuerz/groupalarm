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
| `mqtt_topic` | Yes | `pager/groupalarm/{id}` | MQTT topic template. `{id}` is replaced with the alarm ID; `{org}` is replaced with the organisation ID (legacy) |
| `groupalarm_user_id` | No | `0` | Your GroupAlarm user ID. When set, publishes your personal response state to `…/feedback/own` |

## MQTT Topics

Each alarm is published across multiple topics. `{id}` = alarm ID (e.g. `998877`).

| Topic | Payload | Published |
|-------|---------|-----------|
| `pager/groupalarm/{id}` | Raw alarm message string | Once, on first alarm detection |
| `pager/groupalarm/{id}/title` | Event name (e.g. `B2 Wohnungsbrand`) | Once, on first alarm detection |
| `pager/groupalarm/{id}/feedback/positive` | Integer — confirmed responses | Every 5 s while alarm is open |
| `pager/groupalarm/{id}/feedback/negative` | Integer — declined responses | Every 5 s while alarm is open |
| `pager/groupalarm/{id}/feedback/unknown` | Integer — pending responses | Every 5 s while alarm is open |
| `pager/groupalarm/{id}/feedback/own` | Your personal state: `positive`, `negative`, or `unknown` | Every 5 s (only when `groupalarm_user_id` is set) |
| `pager/groupalarm/{id}/meta` | JSON with all fields (see below) | Every 5 s while alarm is open |

**Example `meta` payload:**
```json
{
  "alarm_id": 998877,
  "organization_id": 19184,
  "event_id": 55432,
  "title": "B2 Wohnungsbrand",
  "message": "Alarm: Feuer in Gebäude A",
  "feedback": { "positive": 3, "negative": 1, "unknown": 5 },
  "feedback_pct": { "positive": 33.3, "negative": 11.1, "unknown": 55.6 }
}
```

## Home Assistant Sensor Examples

```yaml
mqtt:
  sensor:
    - name: "Einsatz Nachricht"
      state_topic: "pager/groupalarm/+/meta"
      value_template: "{{ value_json.message }}"

    - name: "Einsatz Titel"
      state_topic: "pager/groupalarm/+/title"

    - name: "Einsatz Zugesagt"
      state_topic: "pager/groupalarm/+/feedback/positive"
      unit_of_measurement: "Personen"

    - name: "Einsatz Abgelehnt"
      state_topic: "pager/groupalarm/+/feedback/negative"
      unit_of_measurement: "Personen"

    - name: "Einsatz Ausstehend"
      state_topic: "pager/groupalarm/+/feedback/unknown"
      unit_of_measurement: "Personen"

    - name: "Meine Rückmeldung"
      state_topic: "pager/groupalarm/+/feedback/own"
      # state is: positive / negative / unknown
```

To stop an automation when you have responded, use an MQTT trigger on `pager/groupalarm/+/feedback/own` and check `trigger.payload` for `positive` or `negative`.

Use `pager/groupalarm/+` as a trigger topic in automations to fire on any new alarm message.

## Migration from v1.x

The default topic template changed from `pager/groupalarm/{org}` to `pager/groupalarm/{id}`.

If you have existing automations that listen on `pager/groupalarm/{org}` (e.g. `pager/groupalarm/19184`), update them to use `pager/groupalarm/+` as a wildcard trigger, or set `mqtt_topic` back to `pager/groupalarm/{org}` manually (the `{org}` placeholder is still supported).

## Notes

- Alarm IDs are tracked in memory only. Restarting the add-on will cause already-seen alarms from open events to be re-published.
- The add-on connects to the built-in Mosquitto broker by default (`core-mosquitto`). Install the **Mosquitto broker** add-on if you haven't already.
- Feedback counters (`feedback/positive`, etc.) are updated on every poll cycle (every 5 s) as long as the event remains open.
