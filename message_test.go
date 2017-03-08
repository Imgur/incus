package incus

import (
	"encoding/json"
	"testing"

	"github.com/alexjlockwood/gcm"
	apns "github.com/anachronistic/apns"
	mock "github.com/stretchr/testify/mock"
)

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

	mockAPNS := &apns.MockClient{}

	mockAPNS.On("Send", mock.AnythingOfType("*apns.PushNotification")).Return(&apns.PushNotificationResponse{
		Success:       true,
		AppleResponse: "Hello from California!",
		Error:         nil,
	})

	server := &Server{
		Stats:        &DiscardStats{},
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

type MockGCMClient struct {
	mock.Mock
}

func (m *MockGCMClient) Send(msg *gcm.Message, retries int) (resp *gcm.Response, err error) {
	args := m.Called(msg, retries)
	return args.Get(0).(*gcm.Response), args.Error(1)
}

func TestGCM(t *testing.T) {
	msg := new(CommandMsg)
	json.Unmarshal([]byte(`{
		"command": {
			"command": "push",
			"push_type": "android",
			"registration_ids": "123456,654321"
		},
		"message": {
			"event": "foobaz",
			"data": {
				"foobar": "foo"
			},
			"time": 1234
		}
	}`), &msg)

	mockGCM := &MockGCMClient{}

	mockGCM.On("Send", mock.AnythingOfType("*gcm.Message"), mock.AnythingOfType("int")).Return(&gcm.Response{
		MulticastID:  1234,
		Success:      2,
		Failure:      0,
		CanonicalIDs: 0,
		Results: []gcm.Result{
			gcm.Result{MessageID: "abcd", RegistrationID: "123456", Error: ""},
			gcm.Result{MessageID: "bcde", RegistrationID: "654321", Error: ""},
		},
	}, nil)

	server := &Server{
		Stats:       &DiscardStats{},
		gcmProvider: func() GCMClient { return mockGCM },
	}

	msg.FromRedis(server)

	mockGCM.AssertCalled(t, "Send", mock.AnythingOfType("*gcm.Message"), mock.AnythingOfType("int"))

	message := mockGCM.Calls[0].Arguments[0].(*gcm.Message)

	if message.RegistrationIDs == nil || len(message.RegistrationIDs) != 2 {
		t.Fatalf("Expected there to be two registration IDs, instead %+v in %+v", message.RegistrationIDs, message)
	}
}
