package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/term"
	"io"
	"math"
	"net/http"
	"strings"
	"syscall"
)

const FailedAttempt = "no SID cookie found in response header"

type LoginCommand struct {
	Address  string `required:"" help:"the Netgear switch's IP address or host name to connect to" short:"a"`
	Password string `optional:"" help:"the admin console's password; if omitted, it will be prompted for" short:"p"`
}

func (login *LoginCommand) Run(args *GlobalOptions) error {
	if len(login.Password) < 1 {
		pwd, err := promptForPassword(login.Address)
		if err != nil {
			return err
		}
		login.Password = pwd
	}

	if len(login.Password) < 1 {
		return errors.New("no password given")
	}

	seedValue, err := getSeedValueFromSwitch(args, login.Address)
	if err != nil {
		return err
	}

	encryptedPwd := encryptPassword(login.Password, seedValue)

	err = doLogin(args, login.Address, encryptedPwd)
	if err != nil {
		return err
	}

	return nil
}

func promptForPassword(serverName string) (string, error) {
	fmt.Printf("Please enter password for '%s' (input hidden) :> ", serverName)
	// the int conversion is required for the windows build to succeed
	password, err := term.ReadPassword(int(syscall.Stdin))
	println()
	return string(password), err
}

func doLogin(args *GlobalOptions, host string, encryptedPwd string) error {
	url := fmt.Sprintf("http://%s/redirect.html", host)
	if args.Verbose {
		println("login attempt: " + url)
	}
	formData := "password=" + encryptedPwd
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(formData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if args.Verbose {
		println(resp.Status)
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	token := getSessionToken(resp)
	if token == FailedAttempt && resp.StatusCode == http.StatusOK {
		return errors.New("login request returned 200 OK, but response did not contain a session token (SID cookie value#). " +
			"this is known behaviour from the switch. please, wait some minutes and tray again later")
	}

	err = storeToken(args, host, token)
	if err != nil {
		return err
	}

	return nil
}

func checkIsLoginRequired(httpResponseBody string) bool {
	return len(httpResponseBody) < 10 || strings.Contains(httpResponseBody, "/login.cgi")
}

func getSessionToken(resp *http.Response) string {
	cookie := resp.Header.Get("Set-Cookie")
	var sessionIdPrefixes = [...]string{
		"SID=",          // GS308EP, GS308EPP
		"gambitCookie=", // GS316EP
	}
	for _, sessionIdPrefix := range sessionIdPrefixes {
		if strings.HasPrefix(cookie, sessionIdPrefix) {
			sidVal := cookie[len(sessionIdPrefix):]
			split := strings.Split(sidVal, ";")
			return split[0]
		}
	}
	return FailedAttempt
}

func getSeedValueFromSwitch(args *GlobalOptions, host string) (string, error) {
	url := fmt.Sprintf("http://%s/wmi/login", host)
	if args.Verbose {
		println("fetch seed value from: " + url)
	}
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	if args.Verbose {
		println(resp.Status)
	}
	defer resp.Body.Close()

	seedValue, err := getSeedValueFromLoginHtml(resp.Body)
	if err != nil {
		return "", err
	}
	return seedValue, nil
}

func getSeedValueFromLoginHtml(reader io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return "", err
	}
	randVal, exists := doc.Find("#rand").First().Attr("value")

	if exists {
		return randVal, nil
	}
	return "", errors.New("random seed value not found in login.cgi response. " +
		"An element with id=rand and an attribute 'value' is expected")
}

// encryptPassword re-implements some logic from Netgear's GS305EP frontend component, see login.js
func encryptPassword(password string, seedValue string) string {
	mergedStr := specialMerge(password, seedValue)
	hash := md5.New()
	io.WriteString(hash, mergedStr)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func specialMerge(password string, seedValue string) string {
	result := strings.Builder{}
	maxLen := int(math.Max(float64(len(password)), float64(len(seedValue))))
	for i := 0; i < maxLen; i++ {
		if i < len(password) {
			result.WriteString(string([]rune(password)[i]))
		}
		if i < len(seedValue) {
			result.WriteString(string([]rune(seedValue)[i]))
		}
	}
	return result.String()
}
