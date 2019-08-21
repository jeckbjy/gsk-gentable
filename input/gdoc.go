package input

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gentable/base"
	"io/ioutil"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const kDefaultAuthFile = "oauth.json"

type GDocAuthConf struct {
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	RedirectURL  string    `json:"redirect_url"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry,omitempty"`
}

func LoadAuthFromFile(filename string) (*GDocAuthConf, error) {
	if filename == "" {
		filename = kDefaultAuthFile
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	c := &GDocAuthConf{}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, err
	}

	return c, nil
}

type GDocBuilder struct {
	service *sheets.Service
}

func (b *GDocBuilder) Type() string {
	return TypeGdoc
}

func (b *GDocBuilder) Load(opts *Options) ([]LoadFunc, error) {
	if err := b.init(opts.Auth); err != nil {
		return nil, err
	}

	gsheets := opts.Sheets
	if len(opts.Sheets) == 0 {
		// 获取全部,需要先查询一次
		rsp, err := b.service.Spreadsheets.Get(opts.Source).Do()
		if err != nil {
			return nil, err
		}

		for _, sheet := range rsp.Sheets {
			gsheets = append(gsheets, sheet.Properties.Title)
		}
	}

	results := make([]LoadFunc, 0)
	rsp, err := b.service.Spreadsheets.Values.BatchGet(opts.Source).Ranges(gsheets...).Do()
	if err != nil {
		return nil, err
	}
	for i, values := range rsp.ValueRanges {
		tmpName := gsheets[i]
		tmpValues := values
		results = append(results, func() (*base.Sheet, error) {
			return b.parseValues(tmpName, tmpValues)
		})
	}

	return results, nil
}

func (b *GDocBuilder) init(auth *GDocAuthConf) error {
	if auth == nil {
		// 默认从本地加载
		var err error
		auth, err = LoadAuthFromFile("oath.json")
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	endpoint := oauth2.Endpoint{
		AuthURL:   "https://accounts.google.com/o/oauth2/auth",
		TokenURL:  "https://oauth2.googleapis.com/token",
		AuthStyle: 0,
	}

	config := oauth2.Config{
		ClientID:     auth.ClientID,
		ClientSecret: auth.ClientSecret,
		RedirectURL:  auth.RedirectURL,
		Endpoint:     endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/spreadsheets.readonly"},
	}

	token := oauth2.Token{
		AccessToken:  auth.AccessToken,
		RefreshToken: auth.RefreshToken,
		TokenType:    auth.TokenType,
		Expiry:       auth.Expiry,
	}

	source := config.TokenSource(context.Background(), &token)
	tok, err := source.Token()
	if err != nil {
		return err
	}

	client := config.Client(ctx, tok)
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	b.service = srv
	return nil
}

func (b *GDocBuilder) parseValues(name string, vr *sheets.ValueRange) (*base.Sheet, error) {
	sheet := &base.Sheet{}
	sheet.Init(name)
	if len(vr.Values) < 2 {
		return nil, errors.New("no enough data")
	}

	column := len(vr.Values[0])
	for i, row := range vr.Values {
		cells := make([]string, 0, column)
		for _, cell := range row {
			data := fmt.Sprint(cell)
			cells = append(cells, data)
		}
		err := sheet.Append(i, cells)
		if err != nil {
			return nil, err
		}
	}

	return sheet, nil
}
