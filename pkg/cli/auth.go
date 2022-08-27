package cli

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/effective-security/porto/pkg/retriable"
	"github.com/effective-security/porto/x/urlutil"
	"github.com/effective-security/porto/xhttp/header"
	"github.com/effective-security/porto/xhttp/httperror"
	"github.com/effective-security/porto/xhttp/marshal"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/certutil"
	"github.com/effective-security/xpki/jwt/oauth2client"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const (
	EnvAuthTokenName = "ABLE_EMAIL_CLIENT_TOKEN"
)

// AuthCmd is the parent for auth command
type AuthCmd struct {
	Login LoginCmd `cmd:"" help:"login to the server"`
}

// LoginCmd starts login to the server
type LoginCmd struct {
	Provider string `kong:"arg" required:"" help:"provider type: microsoft|google"`
	Config   string `required:"" help:"OAuth2 client configuration"`

	NoStore   bool `help:"Specifies to not store token in the local storage"`
	IDToken   bool `help:"Specifies request id_token"`
	NoBrowser bool `help:"Specifies to disable openning browser"`
}

// Run the command
func (a *LoginCmd) Run(ctx *Cli) error {
	// client, err := ctx.HTTPClient()
	// if err != nil {
	// 	return errors.WithStack(err)
	// }

	client := retriable.New()
	client.EnvAuthTokenName = EnvAuthTokenName
	client.StorageFolder = "~/.ableai/email"

	oa2, err := oauth2client.LoadProvider(a.Config)
	if err != nil {
		return errors.WithMessagef(err, "unable to OAuth config")
	}
	oc := oa2.Client(a.Provider)
	if oc == nil {
		return errors.WithMessagef(err, "unsupported provider: %s", a.Provider)
	}
	o := oc.Config()

	conf := &oauth2.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		RedirectURL:  "http://localhost:38988/auth", // o.RedirectURL,
		Scopes:       o.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  o.AuthURL,
			TokenURL: o.TokenURL,
		},
	}

	rt := "token"
	if a.IDToken {
		rt = "id_token"
	}

	startURL := conf.AuthCodeURL(
		"", // no state
		oauth2.SetAuthURLParam("response_mode", "form_post"),
		oauth2.SetAuthURLParam("response_type", rt),
		oauth2.SetAuthURLParam("nonce", certutil.RandomString(8)),
		oauth2.SetAuthURLParam("provider", a.Provider), // required
	)

	var wg sync.WaitGroup
	handler := func(w http.ResponseWriter, r *http.Request) {
		var err error
		isDone := r.URL.Path == "/auth/done"
		defer func() {
			if isDone || err != nil {
				wg.Done()
			}
		}()

		if isDone {
			w.Write([]byte(`<body onload="window.close()">` +
				`<h2>Authenticated! You may close the browser now.</h2>` +
				`</body>`))
			return
		}

		var (
			token     string
			tokenType string
			expiresIn string
			errCode   string
			errDescr  string
		)
		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				marshal.WriteJSON(w, r, httperror.InvalidRequest("unable to parse response body"))
				return
			}

			token = r.Form.Get("access_token")
			tokenType = r.Form.Get("token_type")
			expiresIn = r.Form.Get("expires_in")
			errCode = r.Form.Get("error")
			errDescr = r.Form.Get("error_description")
		} else {
			vals := r.URL.Query()
			token = urlutil.GetValue(vals, "access_token")
			tokenType = urlutil.GetValue(vals, "token_type")
			expiresIn = urlutil.GetValue(vals, "expires_in")
			errCode = urlutil.GetValue(vals, "error")
			errDescr = urlutil.GetValue(vals, "error_description")
		}
		if errCode != "" {
			marshal.WriteJSON(w, r, httperror.New(http.StatusInternalServerError, errCode, errDescr))
			return
		}

		if token == "" {
			err = httperror.InvalidRequest("missing token parameter")
			marshal.WriteJSON(w, r, err)
			return
		}

		w.Header().Set(header.ContentType, header.TextPlain)
		fmt.Fprintf(ctx.Writer(), "Authenticated! You can close the browser now.\n")

		if a.NoStore {
			fmt.Fprintf(w, "\nTo use the token with DPoP server, run:\nexport %s=%s\n",
				client.EnvAuthTokenName, token)
		} else {
			if tokenType == "" {
				tokenType = "Bearer"
			}
			vals := url.Values{
				"access_token": {token},
				"token_type":   {tokenType},
			}
			if expiresIn != "" {
				ux, err := strconv.ParseInt(expiresIn, 10, 64)
				if err == nil {
					exp := time.Now().Add(time.Duration(ux) * time.Second).Unix()
					vals["exp"] = []string{strconv.FormatInt(exp, 10)}
				}
			}

			err := client.StoreAuthToken(vals.Encode())
			if err != nil {
				fmt.Fprint(ctx.Writer(), err.Error())
			}
		}
		http.Redirect(w, r, "http://localhost:38988/auth/done", http.StatusSeeOther)
	}
	http.HandleFunc("/auth", handler)
	http.HandleFunc("/auth/done", handler)

	wg.Add(1)
	go func() {
		log.Fatal(http.ListenAndServe(":38988", nil))
	}()

	if a.NoBrowser {
		fmt.Fprintf(ctx.Writer(), "open auth URL in browser:\n%s\n", startURL)
	} else {
		execCommand := "xdg-open"

		uname, err := genInfo()
		if err == nil && strings.Contains(uname.Release, "WSL") {
			execCommand = "wsl-open"
		} else if runtime.GOOS == "darwin" {
			execCommand = "open"
		}
		logger.KV(xlog.DEBUG, "open", execCommand, "url", startURL, "runtime", runtime.GOOS, "uname", uname)
		err = exec.Command(execCommand, startURL).Start()
		if err != nil {
			return errors.WithStack(err)
		}
		wg.Wait()
	}

	return nil
}
