package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-resty/resty/v2"
	"gopkg.in/ini.v1"
)

const (
	apiBase      = "https://app.groupalarm.com/api/v1"
	pollInterval = 5 * time.Second
	mqttTimeout  = 10
)

var mqttConnected bool

type Event []struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	OrganizationID int    `json:"organizationID"`
	Archived       bool   `json:"archived"`
}

type Alarm struct {
	Alarms []struct {
		ID      int    `json:"id"`
		Message string `json:"message"`
	} `json:"alarms"`
}

func main() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	apiKey := cfg.Section("groupalarm").Key("api_key").String()
	organisations := cfg.Section("groupalarm").Key("organisations").Strings(",")

	mqttHost := cfg.Section("mqtt").Key("host").String()
	mqttPort, err := cfg.Section("mqtt").Key("port").Int()
	if err != nil {
		log.Fatalf("invalid mqtt port: %v", err)
	}
	mqttUser := cfg.Section("mqtt").Key("user").String()
	mqttPassword := cfg.Section("mqtt").Key("password").String()
	mqttClientID := cfg.Section("mqtt").Key("client").String()
	mqttTopic := cfg.Section("mqtt").Key("topic").String()

	httpClient := resty.New()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", mqttHost, mqttPort))
	opts.SetClientID(mqttClientID)
	opts.SetUsername(mqttUser)
	opts.SetPassword(mqttPassword)
	opts.AutoReconnect = true
	opts.OnConnect = func(_ mqtt.Client) {
		log.Println("[mqtt] connected")
		mqttConnected = true
	}
	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Printf("[mqtt] connection lost: %v", err)
		mqttConnected = false
	}

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("[mqtt] connect failed: %v", token.Error())
	}

	for i := 0; i < mqttTimeout && !mqttConnected; i++ {
		log.Println("[mqtt] waiting for connection...")
		time.Sleep(time.Second)
	}
	if !mqttConnected {
		log.Fatal("[mqtt] timed out waiting for connection")
	}

	knownIDs := make(map[int]struct{})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	log.Println("starting poll loop")

	for {
		select {
		case <-stop:
			log.Println("shutting down")
			mqttClient.Disconnect(250)
			return
		case <-ticker.C:
			for _, org := range organisations {
				pollOrg(httpClient, apiKey, org, mqttClient, mqttTopic, knownIDs)
			}
		}
	}
}

func pollOrg(httpClient *resty.Client, apiKey, org string, mqttClient mqtt.Client, topicTpl string, knownIDs map[int]struct{}) {
	var events Event
	resp, err := httpClient.R().
		SetQueryParam("organization", org).
		SetResult(&events).
		SetHeader("Personal-Access-Token", apiKey).
		Get(apiBase + "/events/open")
	if err != nil {
		log.Printf("[http] failed to fetch events for org %s: %v", org, err)
		return
	}
	if resp.StatusCode() != 200 {
		log.Printf("[http] events: unexpected status %d for org %s", resp.StatusCode(), org)
		return
	}

	for _, ev := range events {
		var alarmResp Alarm
		resp, err = httpClient.R().
			SetQueryParams(map[string]string{
				"organization": strconv.Itoa(ev.OrganizationID),
				"event":        strconv.Itoa(ev.ID),
			}).
			SetResult(&alarmResp).
			SetHeader("Personal-Access-Token", apiKey).
			Get(apiBase + "/alarms")
		if err != nil {
			log.Printf("[http] failed to fetch alarms for event %d: %v", ev.ID, err)
			continue
		}
		if resp.StatusCode() != 200 {
			log.Printf("[http] alarms: unexpected status %d for event %d", resp.StatusCode(), ev.ID)
			continue
		}

		for _, alarm := range alarmResp.Alarms {
			if _, seen := knownIDs[alarm.ID]; seen {
				continue
			}
			if !mqttConnected {
				log.Printf("[mqtt] skipping alarm %d - not connected", alarm.ID)
				continue
			}
			topic := strings.ReplaceAll(topicTpl, "{org}", org)
			mqttClient.Publish(topic, 1, false, alarm.Message)
			knownIDs[alarm.ID] = struct{}{}
			log.Printf("[mqtt] published alarm %d: %s", alarm.ID, alarm.Message)
		}
	}
}
