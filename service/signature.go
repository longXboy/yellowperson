package service

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"sort"
	"strings"
	"time"
	"yellowPerson/schema"
	"yellowPerson/util"
	"yellowPerson/util/config"
	"yellowPerson/util/db"
	"yellowPerson/util/log"
)

func CreateSignature(request *restful.Request, response *restful.Response) {
	content, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		log.ERROR.Printf("Read body data failed!err:=%v", err)
		return
	}
	r := request.Request
	request.Request.ParseForm()
	var msg_signature string = strings.Join(r.Form["msg_signature"], "")
	var timestamp string = strings.Join(r.Form["timestamp"], "")
	var nonce string = strings.Join(r.Form["nonce"], "")
	if !util.WxCheckValidate(timestamp, nonce, content, msg_signature) {
		log.ERROR.Printf("Invalid data(%s) failed!err:=%v", string(content), err)
		return
	}
	data, err := util.WxXmlDecode(content)
	if err != nil {
		log.ERROR.Printf("WxXmlDecode data failed!err:=%v", err)
		return
	}
	passcode, err := config.GetConfigString("manage.passcode")
	if err != nil {
		log.ERROR.Printf("GetConfigString manage.passcode failed!err:=%v", err)
		return
	}
	log.INFO.Println("dataConent:", data.Content, "openId:", data.FromUserName)
	if passcode != util.HashPassword(data.Content) {
		fmt.Printf("psw not matched!raw:=%s,hashedcontent=%s", data.Content, util.HashPassword(data.Content))
		return
	}
	times, err := config.GetConfigInt("manage.passcode.times")
	if err != nil {
		log.ERROR.Printf("GetConfigString manage.passcode.times failed!err:=%v", err)
		return
	}
	duration, err := config.GetConfigInt("manage.passcode.duration")
	if err != nil {
		log.ERROR.Printf("GetConfigString manage.passcode.duration failed!err:=%v", err)
		return
	}
	var manage schema.Manage
	IsOK, err := db.DB.Where("OpenId=?", data.FromUserName).Get(&manage)
	if err != nil {
		log.ERROR.Printf("Found OpenId(%s) from Db failed!err:=%v", data.FromUserName, err)
		return
	}
	if !IsOK {
		log.ERROR.Printf("OpenId(%s) Not found !err:=%v", data.FromUserName, err)
		return
	}
	manage.AccessCode = util.HashPassword(fmt.Sprintf("%d", time.Now().UnixNano()))[:16]
	manage.Times = times
	manage.ExpiredTs = int(time.Now().Unix()) + duration
	rows, err := db.DB.Where("OpenId=?", data.FromUserName).Update(&manage)
	if err != nil || rows < 1 {
		log.ERROR.Printf("Update Accesscode failed where OpenId=%s!err:=%v", data.FromUserName, err)
		return
	}
	dataTo := util.MsgRawTo{}
	dataTo.Content = manage.AccessCode
	dataTo.CreateTime = int(time.Now().Unix())
	dataTo.FromUserName = data.ToUserName
	dataTo.MsgType = data.MsgType
	dataTo.ToUserName = data.FromUserName
	dataTo.XMLName = data.XMLName
	Reply, err := util.WxXmlEncode(dataTo)
	if err != nil {
		log.ERROR.Printf("WxXmlEncode data(%v) failed!err:=%v", dataTo, err)
		return
	}
	fmt.Fprintf(response.ResponseWriter, string(Reply))
	fmt.Println(string(Reply))
}

func CheckSignature(req *restful.Request, resp *restful.Response) {
	r := req.Request
	req.Request.ParseForm()
	var token string = "mkwhatyellowpersonlongxia"
	var signature string = strings.Join(r.Form["signature"], "")
	var timestamp string = strings.Join(r.Form["timestamp"], "")
	var nonce string = strings.Join(r.Form["nonce"], "")
	var echostr string = strings.Join(r.Form["echostr"], "")
	tmps := []string{token, timestamp, nonce}
	sort.Strings(tmps)
	tmpStr := tmps[0] + tmps[1] + tmps[2]
	tmp := util.Str2sha1(tmpStr)
	if tmp == signature {
		fmt.Fprintf(resp.ResponseWriter, echostr)
	} else {
		fmt.Fprintf(resp.ResponseWriter, "Invalid")
	}

}
