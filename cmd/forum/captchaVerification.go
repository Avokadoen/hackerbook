package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)


// ValidateReCaptcha sends the response token from reCaptcha on signup page, and
// sends it to Google's API for verification.
func ValidateReCaptcha(token string) bool{

	form := url.Values{}

	form.Add("secret", os.Getenv("CAPTCHASECRET"))
	form.Add("response", token)
	req, err := http.NewRequest("POST", "https://www.google.com/recaptcha/api/siteverify", strings.NewReader(form.Encode()))

	if err != nil {
		fmt.Printf("Unable to make request, err: %+v", err)
		return false
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Unable to do request, err: %+v", err)
		return false
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Failed to read body, err: %+v", err)
		return false
	}

	response := ReCaptchaResponse{}

	err = json.Unmarshal(body, &response)

	if len(response.Errorcode) > 0 {
		for _, element := range response.Errorcode {
			fmt.Printf("Errorcode: %v", element)
		}
		return false
	}

	if err != nil {
		fmt.Printf("Unable to unmarshal response, err: %+v", err)
		return false
	}

	return response.Success
}
