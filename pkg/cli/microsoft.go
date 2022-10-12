package cli

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/effective-security/porto/xhttp/header"
	"github.com/effective-security/porto/xhttp/httperror"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/jwt"
	"github.com/effective-security/xpki/jwt/oauth2client"
)

func getMicrosoftUser(ctx context.Context, cl *oauth2client.ClientConfig, accessToken string) (jwt.MapClaims, error) {
	req, _ := http.NewRequest(http.MethodGet, cl.UserinfoURL, nil)
	req.Header.Add(header.Authorization, "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, httperror.Unexpected("unable to get user info").WithCause(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, httperror.Unexpected("unable to decode user info").WithCause(err)
	}
	logger.KV(xlog.DEBUG, "userinfo", string(body))

	var res jwt.MapClaims
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, httperror.Unexpected("unable to decode user info").WithCause(err)
	}

	return res, nil
}
