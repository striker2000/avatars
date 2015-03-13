package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const googleAPIKey = "AIzaSyBtpxcQznzxFY6tsYV9O2DbChhegTtMWvs"

func main() {
	http.HandleFunc("/", getAvatar)

	log.Fatal(http.ListenAndServe(":3000", nil))
}

func getAvatar(w http.ResponseWriter, req *http.Request) {
	parts := strings.Split(req.URL.Path, "/")
	if len(parts) != 3 || len(parts[2]) == 0 {
		http.NotFound(w, req)
		return
	}

	network, userID := parts[1], parts[2]

	var avatarURL string
	var err error

	switch network {
	case "fb":
		avatarURL, err = getFbAvatar(userID)
	case "google":
		avatarURL, err = getGoogleAvatar(userID)
	case "gravatar":
		avatarURL, err = getGravatarAvatar(userID)
	case "vk":
		avatarURL, err = getVkAvatar(userID)
	}

	if err != nil {
		log.Print(err)
		http.Error(w, "Internal error", 500)
		return
	}

	if len(avatarURL) == 0 {
		http.NotFound(w, req)
		return
	}

	//fmt.Fprint(w, avatarURL)
	http.Redirect(w, req, avatarURL, 303)
}

type fbResponse struct {
	Data struct {
		URL string
	}
}

func getFbAvatar(userID string) (string, error) {
	url := fmt.Sprintf(
		"http://graph.facebook.com/%s/picture?redirect=false&type=large",
		url.QueryEscape(userID),
	)
	var payload fbResponse

	err := getJSON(url, &payload)
	if err != nil {
		return "", err
	}

	return payload.Data.URL, nil
}

type googleResponse struct {
	Image struct {
		URL string
	}
}

func getGoogleAvatar(userID string) (string, error) {
	url := fmt.Sprintf(
		"https://www.googleapis.com/plus/v1/people/%s?fields=image&key=%s",
		url.QueryEscape(userID),
		url.QueryEscape(googleAPIKey),
	)
	var payload googleResponse

	err := getJSON(url, &payload)
	if err != nil {
		return "", err
	}

	return payload.Image.URL, nil
}

func getGravatarAvatar(userID string) (string, error) {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(userID)))

	return fmt.Sprintf(
		"http://www.gravatar.com/avatar/%s",
		url.QueryEscape(hash),
	), nil
}

type vkResponse struct {
	Response []struct {
		PhotoMax string `json:"photo_max"`
	}
}

func getVkAvatar(userID string) (string, error) {
	url := fmt.Sprintf(
		"https://api.vk.com/method/users.get?user_ids=%s&fields=photo_max",
		url.QueryEscape(userID),
	)
	var payload vkResponse

	err := getJSON(url, &payload)
	if err != nil {
		return "", err
	}

	if len(payload.Response) == 0 {
		return "", nil
	}

	return payload.Response[0].PhotoMax, nil
}

func getJSON(url string, payload interface{}) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		if res.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("Bad response status: %d", res.StatusCode)
	}

	return json.NewDecoder(res.Body).Decode(payload)
}
