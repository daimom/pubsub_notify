// Package helloworld provides a set of Cloud Functions samples.
package helloworld

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"log"
	"net/http"
)

func init() {
	functions.CloudEvent("HelloPubSub", helloPubSub)
	functions.CloudEvent("sendDiscord", sendDiscord)
}

// MessagePublishedData contains the full Pub/Sub message
// See the documentation for more details:
// https://cloud.google.com/eventarc/docs/cloudevents#pubsub
type MessagePublishedData struct {
	Message PubSubMessage
}

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type PubSubMessage struct {
	Data       []byte            `json:"data"`
	Attributes map[string]string `json:"attributes"`
}
type Payload struct {
	ResourceType       string `json:"resourceType"`
	Operation          string `json:"operation"`
	OperationStartTime string `json:"operationStartTime"`
	CurrentVersion     string `json:"currentVersion"`
	TargetVersion      string `json:"targetVersion"`
}

func sendDiscord(ctx context.Context, e event.Event) error {
	var msg MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		return fmt.Errorf("event.DataAs: %v", err)
	}
	attr := msg.Message.Attributes
	log.Printf("All attr:%s", attr)
	text := string(msg.Message.Data) // Automatically decoded from base64.
	if text == "" {
		text = "World"
	}
	log.Printf("%s!", text)

	// 印出 attributes 的內容
	for key, value := range attr {
		text = text + "\n" + key + ":" + value
		if key == "payload" {
			payloadJSON := attr["payload"]
			payload := Payload{}
			err := json.Unmarshal([]byte(payloadJSON), &payload)
			if err != nil {
				fmt.Println("解析 payload 時發生錯誤:", err)
				continue
			}
			text = text + "\n\t" + "ResourceType:" + payload.ResourceType
			text = text + "\n\t" + "Operation:" + payload.Operation
			text = text + "\n\t" + "OperationStartTime:" + payload.OperationStartTime
			text = text + "\n\t" + "CurrentVersion:" + payload.CurrentVersion
			text = text + "\n\t" + "TargetVersion:" + payload.TargetVersion
			continue
		}

	}

	// 定義 Webhook URL
	webhookURL := "https://URL"

	// 建立 payload
	payload := map[string]string{
		"content": text,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("建立 payload 時發生錯誤:%s", err)
		return nil
	}

	// 建立 POST 請求
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("傳送 POST 請求時發生錯誤:%s", err)
		return nil
	}
	defer resp.Body.Close()

	// 檢查回應狀態碼
	if resp.StatusCode == http.StatusNoContent {
		log.Printf("訊息已成功發送！")
	} else {
		log.Printf("訊息發送失敗，狀態碼:%s", resp.StatusCode)
	}
	return nil
}
