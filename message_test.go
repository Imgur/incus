package incus

import "encoding/json"
import "testing"
import apns "github.com/anachronistic/apns"
import mock "github.com/stretchr/testify/mock"

func TestAPNS(t *testing.T) {
	msg := new(CommandMsg)
	json.Unmarshal([]byte(`{
		"command": {
			"command": "push",
			"push_type": "ios",
			"build": "store",
			"device_token": "123456"
		},
		"message": {
			"event": "foobaz",
			"data": {
				"message_text": "foobar"
			},
			"time": 1234
		}
	}`), &msg)

	configVars := make(map[string]string)
	configVars["ios_push_sound"] = "bingbong.wav"

	mockAPNS := &apns.MockClient{}

	mockAPNS.On("Send", mock.AnythingOfType("*apns.PushNotification")).Return(&apns.PushNotificationResponse{
		Success:       true,
		AppleResponse: "Hello from California!",
		Error:         nil,
	})

	server := &Server{
		Stats: &DiscardStats{},
		Config: &Configuration{
			vars: configVars,
		},
		apnsProvider: func(build string) apns.APNSClient { return mockAPNS },
	}

	msg.FromRedis(server)

	mockAPNS.AssertCalled(t, "Send", mock.AnythingOfType("*apns.PushNotification"))

	pushNotification := mockAPNS.Calls[0].Arguments[0].(*apns.PushNotification)

	if pushNotification.DeviceToken != "123456" {
		t.Fatalf("Expected device token to be 123456, instead %s in %+v", pushNotification.DeviceToken, pushNotification)
	}

	apsPayload := pushNotification.Get("aps").(*apns.Payload)

	if apsPayload.Alert != "foobar" {
		t.Fatalf("Expected push alert to be \"foobar\", instead %s in %+v", apsPayload.Alert, pushNotification)
	}
}
