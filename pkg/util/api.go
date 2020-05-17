package util

import (
	"fmt"
	"net/http"
)

func MakeGetRequest(bearerTok string, url string) (*http.Response, error) {
	var bearer = fmt.Sprintf("Bearer %s", bearerTok)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", bearer)
	client := &http.Client{}
	return client.Do(req)
}
