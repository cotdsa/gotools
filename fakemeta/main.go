package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/vault/api"
	"net/http"
	"time"
)

type IAMResponseType struct {
	Code            string
	LastUpdated     string
	Type            string
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      string
}

type CacheType struct {
	UserToken         string
	UserTokenExpiry   time.Time
	IAMResponse       IAMResponseType
	IAMResponseExpiry time.Time
}

type appContext struct {
	currentProfile string
	currentPath    string
	cache          CacheType
}

func authenticate_userpass(c *api.Client, user string, pass string) (string, time.Time, error) {
	var (
		expiry time.Time
		token  string
		err    error
		resp   *api.Response
	)

	bodystruct := struct {
		Password string `json:"password"`
	}{
		Password: pass,
	}

	req := c.NewRequest("POST", "/v1/auth/userpass/login/"+user)
	req.SetJSONBody(bodystruct)

	resp, err = c.RawRequest(req)

	if err != nil {
		return token, expiry, err
	}

	var output interface{}
	jd := json.NewDecoder(resp.Body)
	err = jd.Decode(&output)

	if err != nil {
		return token, expiry, err
	}

	body := output.(map[string]interface{})
	auth := body["auth"].(map[string]interface{})
	token = auth["client_token"].(string)

	//expiry = time.Now().Add(time.Second * time.Duration(auth["lease_duration"].(float64)))
	expiry = time.Now().Add(time.Second * time.Duration(30))

	return token, expiry, nil
}

func get_vault_iam_creds(c *api.Client, path string) (string, string, string, time.Time, error) {
	var (
		access_key     string
		secret_key     string
		security_token string
		expiry         time.Time
		err            error
	)

	l := c.Logical()
	s, err := l.Read(path)
	if err != nil {
		return access_key, secret_key, security_token, expiry, err
	}

	access_key = s.Data["access_key"].(string)
	secret_key = s.Data["secret_key"].(string)
	security_token = s.Data["security_token"].(string)
	//expiry = time.Now().Add(time.Second * time.Duration(s.LeaseDuration))
	expiry = time.Now().Add(time.Second * time.Duration(10))

	return access_key, secret_key, security_token, expiry, nil

}

func (ah *appContext) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var (
		access_key     string
		secret_key     string
		security_token string
		err            error
	)

	w.Header().Set("Content-Type", "application/json")

	requested_profile := r.URL.Path[len("/latest/meta-data/iam/security-credentials/"):]

	if requested_profile == "" {
		fmt.Fprintf(w, "%s", ah.currentProfile)
		return
	}

	//f, _ := w.(http.Flusher)
	//f.Flush()

	config := api.DefaultConfig()
	config.Address = "https://vault.cgws.com.au"
	c, _ := api.NewClient(config)

	if time.Now().After(ah.cache.UserTokenExpiry) {

		vault_username, vault_password, err := get_user_password_interactive()
		if err != nil {
			fmt.Fprintf(w, "%v", err)
			return
		}

		ah.cache.UserToken, ah.cache.UserTokenExpiry, err = authenticate_userpass(c, vault_username, vault_password)
		if err != nil {
			fmt.Fprintf(w, "%v", err)
			return
		}
	}
	c.SetToken(ah.cache.UserToken)

	if time.Now().After(ah.cache.IAMResponseExpiry) {
		access_key, secret_key, security_token, ah.cache.IAMResponseExpiry, err = get_vault_iam_creds(c, ah.currentPath+"powerusersts")

		ah.cache.IAMResponse = IAMResponseType{
			Code:            "Success",
			LastUpdated:     time.Now().Format("2006-01-02T15:04:05Z"),
			Type:            "AWS-HMAC",
			AccessKeyId:     access_key,
			SecretAccessKey: secret_key,
			Token:           security_token,
			Expiration:      ah.cache.IAMResponseExpiry.Format("2006-01-02T15:04:05Z"),
		}
	}

	js, err := json.Marshal(ah.cache.IAMResponse)
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}

	w.Write(js)
}

func main() {

	context := &appContext{
		currentProfile: "s3access",
		currentPath:    "awstest/sts/",
		cache:          CacheType{},
	}

	http.Handle("/latest/meta-data/iam/security-credentials/", context)
	err := http.ListenAndServe("169.254.169.254:80", nil)
	fmt.Println(err)
}
