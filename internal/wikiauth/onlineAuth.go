package wikiauth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ktt-ol/sesam/internal/conf"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Auth based on https://github.com/smilix/moinAuthProvider

// without the concrete action
const pageParam = "/?action=authService&do="
const maxCacheAge = time.Duration(24 * time.Hour)

type onlineAuth struct {
	log                 *logrus.Entry
	wikiActionUrl       string
	authToken           string
	nameToEmailMapCache map[string]string
	lastCacheUpdate     time.Time
	updateCacheMux      sync.Mutex
}

func NewOnlineAuth(config *conf.AuthOnline) WikiAuth {
	url := config.WikiBaseUrl + pageParam

	auth := onlineAuth{
		log:                 logrus.WithField("where", "onlineAuth"),
		wikiActionUrl:       url,
		authToken:           config.AuthToken,
		nameToEmailMapCache: make(map[string]string),
		lastCacheUpdate:     time.Unix(0, 0),
	}

	return &auth
}

func (a *onlineAuth) CheckPassword(emailOrName string, password string) (userName string, authError *AuthError) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
			authError = &AuthError{Error: err, SystemError: true}
		}
	}()

	userName = emailOrName
	if strings.Contains(emailOrName, "@") {
		name, notFound := a.getNameForEmail(emailOrName)
		if notFound {
			return "", &AuthError{Error: errors.New("email doesn't exist"), LoginNotFound: true}
		}
		userName = name
	}

	result := a.requestLogin(userName, password)
	if result == "ok" {
		return userName, nil
	} else if result == "unknown_user" {
		return userName, &AuthError{Error: errors.New("name doesn't exist (or not in Member group)"), LoginNotFound: true}
	} else if result == "wrong_password" {
		return userName, &AuthError{Error: errors.New("Invalid password")}
	} else {
		return "", &AuthError{Error: errors.New("Unexpected result: " + result), SystemError: true}
	}
}

func (a *onlineAuth) getNameForEmail(email string) (name string, emailNotFound bool) {
	if a.lastCacheUpdate.Add(maxCacheAge).Before(time.Now()) {
		a.updateUserList()
	}

	email, found := a.nameToEmailMapCache[strings.ToLower(email)]
	if !found {
		return "", true
	}
	return email, false
}

func (a *onlineAuth) updateUserList() {
	a.updateCacheMux.Lock()
	defer a.updateCacheMux.Unlock()

	// check again, because another thread might have updated the cache in between
	if !a.lastCacheUpdate.Add(maxCacheAge).Before(time.Now()) {
		return
	}

	a.log.Info("Updating user list")

	userList := a.requestUserList()
	newCache := make(map[string]string)
	for _, user := range userList {
		newCache[strings.ToLower(user.Email)] = user.Login
	}

	a.lastCacheUpdate = time.Now()
	a.nameToEmailMapCache = newCache
}

func (a *onlineAuth) requestUserList() []response_user {
	req, err := http.NewRequest("POST", a.wikiActionUrl+"list", nil)
	panicForError(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-Token", a.authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	panicForError(err)

	body, err := ioutil.ReadAll(resp.Body)
	panicForError(err)
	defer resp.Body.Close()

	var result []response_user
	err = json.Unmarshal(body, &result);
	if err != nil {
		a.log.WithField("http", resp.StatusCode).WithField("data", abbr(string(body), 20)).
			WithError(err).Error("Invalid json for user response")
		panic(err)
	}

	return result
}

func (a *onlineAuth) requestLogin(username string, password string) string {
	message := map[string]interface{}{
		"login":    username,
		"password": password,
	}
	bytesRepresentation, err := json.Marshal(message)
	panicForError(err)

	req, err := http.NewRequest("POST", a.wikiActionUrl+"loginCheck", bytes.NewBuffer(bytesRepresentation))
	panicForError(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-Token", a.authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	panicForError(err)

	body, err := ioutil.ReadAll(resp.Body)
	panicForError(err)
	defer resp.Body.Close()

	var result response_loginResult
	err = json.Unmarshal(body, &result);
	if err != nil {
		a.log.WithField("http", resp.StatusCode).WithField("data", abbr(string(body), 20)).
			WithError(err).Error("Invalid json for login")
		panic(err)
	}

	return result.Result
}

func abbr(msg string, maxSize int) string {
	if maxSize <= 0 || len(msg) <= maxSize {
		return msg
	}

	return msg[0:maxSize] + "..."
}

func panicForError(err error) {
	if err != nil {
		panic(err)
	}
}

type response_user struct {
	Login string
	Email string
}

type response_loginResult struct {
	Result string
}
