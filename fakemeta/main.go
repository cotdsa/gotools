package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/vault/api"
	"net/http"
	"time"
	//"os"
)

type IAMResponse struct {
	Code            string
	LastUpdated     string
	Type            string
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      string
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

	expiry = time.Now().Local().Add(time.Second * time.Duration(auth["lease_duration"].(float64)))

	return token, expiry, nil
}

func get_iam_creds(c *api.Client, rolename string) (string, string, string, time.Time, error) {
	var (
		access_key     string
		secret_key     string
		security_token string
		expiry         time.Time
		err            error
	)

	l := c.Logical()
	s, err := l.Read(fmt.Sprintf("awstest/sts/%s", rolename))
	if err != nil {
		return access_key, secret_key, security_token, expiry, err
	}
	for key, value := range s.Data {
		fmt.Printf("%s\t%v\n", key, value)
	}
	access_key = s.Data["access_key"].(string)
	secret_key = s.Data["secret_key"].(string)
	security_token = s.Data["security_token"].(string)
	expiry = time.Now().Local().Add(time.Second * time.Duration(s.LeaseDuration))

	return access_key, secret_key, security_token, expiry, nil

}

func handler(w http.ResponseWriter, r *http.Request) {

	config := api.DefaultConfig()
	config.Address = "https://vault.cgws.com.au"
	c, _ := api.NewClient(config)

	//if err != nil {
	//	panic(err)
	//}
	vault_username, vault_password, err := get_user_password_interactive()
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}

	token, _, err := authenticate_userpass(c, vault_username, vault_password)
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}

	c.SetToken(token)

	access_key, secret_key, security_token, exipry, err := get_iam_creds(c, "powerusersts")

	iamresponse := IAMResponse{
		Code:            "Success",
		LastUpdated:     time.Now().Local().Format("2006-01-02T15:04:05Z"),
		Type:            "AWS-HMAC",
		AccessKeyId:     access_key,
		SecretAccessKey: secret_key,
		Token:           security_token,
		Expiration:      exipry.Format("2006-01-02T15:04:05Z"),
	}

	js, err := json.Marshal(iamresponse)
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	//fmt.Fprintf(w, "Hi there %s, Your vault token is %s and will expire at %s", vault_username, token, expiry)
	//os.Exit(0)
}

func main() {
	t := time.Now()
	fmt.Println(t.Format("2012-04-26T16:39:16Z"))
	http.HandleFunc("/latest/meta-data/iam/security-credentials/s3access", handler)
	//a, b, err := get_user_password_interactive()
	//fmt.Printf("user=%s, pass=%s", a, b)
	err := http.ListenAndServe("169.254.169.254:80", nil)
	fmt.Println(err)
}
