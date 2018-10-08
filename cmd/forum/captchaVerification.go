package main

import (
	"os"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"strings"
	"net/url"
)

// ValidateReCaptcha sends the response token from reCaptcha on signup page, and
// sends it to Google's API for verification.
func ValidateReCaptcha(token string) bool{

	form := url.Values{}

	form.Add("secret", os.Getenv("CAPTCHASECRET"))
	form.Add("response", token)
	req, err := http.NewRequest("POST", "https://www.google.com/recaptcha/api/siteverify", strings.NewReader(form.Encode()))

	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return false
	}

	response := ReCaptchaResponse{}

	err = json.Unmarshal(body, &response)

	if len(response.Errorcode) > 0 {
		return false
	}

	if err != nil{
		return false
	}

	return response.Success
}
