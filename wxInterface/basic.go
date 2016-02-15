package wxInterface

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"net/http"
	"yellowPerson/util/config"
	"yellowPerson/util/log"
)

type AccessToken struct {
	Token    string
	ExpireIn int64
}

type JsApiTicket struct {
	Ticket   string
	ExpireIn int64
}

func GetWeiXinToken() (*AccessToken, error) {
	wxappid, err := config.GetConfigString(config.Key_WeiXinAppId)
	if err != nil {
		return nil, err
	}
	wxappsecret, err := config.GetConfigString(config.Key_WeiXinAppSecret)
	if err != nil {
		return nil, err
	}
	getTokenUrl := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", wxappid, wxappsecret)
	resp, err := http.Get(getTokenUrl)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	sjson, err := simplejson.NewJson(content)
	if err != nil {
		return nil, err
	}
	token, err := sjson.Get("access_token").String()
	if err != nil {
		return nil, err
	}
	expireIn, err := sjson.Get("expires_in").Int64()
	if err != nil {
		return nil, err
	}
	logErrMsg(sjson)
	return &AccessToken{token, expireIn}, nil
}

func GetWeiXinJsApi(acctoken string) (*JsApiTicket, error) {
	getJsApiUrl := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=%s&type=jsapi", acctoken)
	resp, err := http.Get(getJsApiUrl)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	sjson, err := simplejson.NewJson(content)
	if err != nil {
		return nil, err
	}
	ticket, err := sjson.Get("ticket").String()
	if err != nil {
		return nil, err
	}
	expireIn, err := sjson.Get("expires_in").Int64()
	if err != nil {
		return nil, err
	}
	logErrMsg(sjson)
	return &JsApiTicket{ticket, expireIn}, nil
}

func logErrMsg(sjson *simplejson.Json) {
	errcode, err := sjson.Get("errcode").Int()
	if err == nil {
		if errcode != 0 {
			errmsg, err := sjson.Get("errmsg").String()
			if err == nil {
				log.ERROR.Printf("Get ApiJsTicketFailed!errmsg=%v,errorcode=%d", errmsg, errcode)
			}
		}
	}
}
