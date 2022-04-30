package client

import (
"fmt"

"github.com/twilio/twilio-go"
openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwilioClient struct {
	Client *twilio.RestClient
}

func NewTwilioClient(accountSid string, authToken string) *TwilioClient {
	return &TwilioClient{
		Client: twilio.NewRestClientWithParams(
			twilio.ClientParams{
				Username: accountSid,
				Password: authToken,
			}),
	}
}

func (tc *TwilioClient) SendText(phoneNumber string, message string) error {
	params := &openapi.CreateMessageParams{}
	//params.SetTo("+15558675309")
	params.SetTo(phoneNumber)
	params.SetFrom("+14706845662")
	params.SetBody(message)

	resp, err := tc.Client.ApiV2010.CreateMessage(params)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println("Message Sid: " + *resp.Sid)
		return nil
	}
}