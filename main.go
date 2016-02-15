package main

import (
	"flag"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/krig/go-sox"
	"html/template"
	slog "log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"yellowPerson/cache"
	"yellowPerson/schema"
	"yellowPerson/service"
	"yellowPerson/util"
	"yellowPerson/util/config"
	"yellowPerson/util/db"
	l "yellowPerson/util/log"
)

var pconfig *string
var psection *string

func initConfig() {
	pconfig = flag.String("config", "./config.cfg", "config file")
	psection = flag.String("sec", "dev", "section of config file to apply")
	flag.Parse()
	env := map[string]string{
		"config": *pconfig,
		"sec":    *psection,
	}
	config.InitConfigEnv(env)
	err := config.LoadConfigFile()
	if err != nil {
		slog.Fatalf("LoadConfigFile from %s failed. err=%v\n", *pconfig, err)
		return
	}
}

func initLog() {
	logpath, err := config.GetConfigString("global.logpath")
	if err != nil {
		slog.Fatalf("load global.logpath failed in config file %s[%s]\n", *pconfig, *psection)
		return
	}
	alogpath, err := config.GetConfigString("global.alogpath")
	if err != nil {
		slog.Fatalf("load global.alogpath failed in config file %s[%s]\n", *pconfig, *psection)
		return
	}
	err = l.InitLogger(logpath, alogpath)
	if err != nil {
		slog.Fatalf("InitLogger with logpath %s failed. err=%v\n", logpath, err)
	}

}

func filter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	beforets := time.Now().UnixNano()
	/*compress, _ := restful.NewCompressingResponseWriter(resp.ResponseWriter, restful.ENCODING_GZIP)
	resp.ResponseWriter = compress
	defer func() {
		compress.Close()
	}()*/
	chain.ProcessFilter(req, resp)
	responsets := (time.Now().UnixNano() - beforets) / 1000000
	var acclog schema.AccLog
	acclog.CreateTs = int(time.Now().Unix())
	acclog.ContentLength = resp.ContentLength()
	acclog.IP = util.GetClientIP(req)
	acclog.Method = req.Request.Method
	acclog.ResponseTs = int(responsets)
	acclog.StatusCode = resp.StatusCode()
	acclog.UA = req.Request.UserAgent()
	acclog.URI = req.Request.URL.RequestURI()
	_, err := db.DB.InsertOne(&acclog)
	if err != nil {
		l.ERROR.Println("Insert acclog into Db Err!,err:=%v", err)
	}
}

func main() {
	ws := new(restful.WebService)
	restful.Filter(filter)
	initConfig()
	cache.InitRedis()
	initLog()
	if !sox.Init() {
		slog.Fatal("Failed to initialize SoX")
	}
	defer sox.Quit()
	err := db.InitDB()
	if err != nil {
		slog.Fatalf("InitDB failed. err=%v\n", err)
	}
	defer db.FiniDB()
	util.Lock = new(sync.Mutex)
	ws.Route(ws.GET("/check").To(service.CheckSignature))
	ws.Route(ws.POST("/check").To(service.CreateSignature))
	ws.Route(ws.GET("/share/minions/{audio-id}").To(getMinions))
	ws.Route(ws.GET("/share/getminonsid").To(getMinionsId).Produces(restful.MIME_JSON))
	ws.Route(ws.GET("/static/{subpath:*}").To(static))
	ws.Route(ws.POST("/share/getmediaid").To(getId).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON))
	restful.Add(ws)
	apilisten, err := config.GetConfigString("api.listen")
	if err != nil {
		apilisten = ":80"
	}
	err = http.ListenAndServe(apilisten, nil)
	if err != nil {
		slog.Fatalf("listen and serve failed!err=%v\n", err)
	}
	defer cache.Close()
	defer l.FiniLogger()
}

func static(req *restful.Request, resp *restful.Response) {
	file, _ := exec.LookPath(os.Args[0])
	rootdir := filepath.Dir(file) + `/static`
	actual := path.Join(rootdir, req.PathParameter("subpath"))
	http.ServeFile(
		resp.ResponseWriter,
		req.Request,
		actual)
}

type MinionsData struct {
	WxConfigs WxConfig
	AudioUrl  string
	MinionsId int64
}

type GetMinionsId struct {
	Id int64 `json:"Id"`
}

type WxConfig struct {
	Debug     bool
	AppId     string
	Timestamp int64
	NonceStr  string
	Signature string
	JsApiList []string
}

func getMinionsId(request *restful.Request, response *restful.Response) {

	var rsp GetMinionsId
	defer response.WriteEntity(&rsp)
	var id schema.Minions
	id.CreateTs = int(time.Now().Unix())
	_, err := db.DB.InsertOne(&id)
	if err != nil {
		l.ERROR.Printf("Insert into Minions failed ,err=%v", err)
		rsp.Id = time.Now().Unix()
		return
	}
	rsp.Id = id.Id
}

func getMinions(request *restful.Request, response *restful.Response) {
	response.AddHeader("Content-Type", "text/html;charset=utf-8")
	id := request.PathParameter("audio-id")
	ticket, err := cache.GetJsApiTicket()
	if err != nil || ticket == "" {
		l.ERROR.Printf("Get JsApiTicket from Redis failed,err=%v,token=%s", err, ticket)
		tmpl, _ := template.ParseFiles("./template/minions_play.html")
		tmpl.Execute(response, nil)
		return
	}
	wxappid, err := config.GetConfigString(config.Key_WeiXinAppId)
	if err != nil {
		l.ERROR.Printf("Get Key_WeiXinAppId failed,err=%v,token=%s", err, ticket)
		tmpl, _ := template.ParseFiles("./template/minions_play.html")
		tmpl.Execute(response, nil)
		return
	}
	timestamp := time.Now().Unix()
	nonceStr := util.Str2sha1(fmt.Sprintf("%d", time.Now().UnixNano()))
	wxconfig := WxConfig{
		Debug:     true,
		AppId:     wxappid,
		Timestamp: timestamp,
		NonceStr:  nonceStr,
	}
	wxconfig.Signature = util.Str2sha1(fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%d&url=%s", ticket, nonceStr, timestamp, "http://"+request.Request.Host+request.Request.RequestURI))
	wxconfig.JsApiList = []string{"startRecord", "stopRecord", "onRecordEnd", "playVoice", "pauseVoice", "stopVoice", "uploadVoice", "downloadVoice", "onMenuShareTimeline", "onMenuShareAppMessage", "translateVoice"}
	tmpl, err := template.ParseFiles("./template/minions_play.html")
	if err != nil {
		l.ERROR.Printf("ParseFile ./template/minions_play.html failed ,err :=%v ", err)
		return
	}
	url := fmt.Sprintf("http://7xlago.com2.z0.glb.qiniucdn.com/voice/%s.mp3", id)
	var mediaLog schema.MediaLog
	IsOK, err := db.DB.Where("`Key`=?", id).Get(&mediaLog)
	if err != nil || !IsOK {
		l.ERROR.Printf("Get mediaLog from db failed ,err :=%v ", err)
		return
	}
	err = tmpl.Execute(response, MinionsData{wxconfig, url, mediaLog.RefId})
	if err != nil {
		l.ERROR.Printf("Execute ./template/minions_play.html failed ,err :=%v ", err)
		return
	}
}

func minions(req *restful.Request, resp *restful.Response) {
	resp.AddHeader("Content-Type", "text/html;charset=utf-8")
	ticket, err := cache.GetJsApiTicket()
	if err != nil || ticket == "" {
		l.ERROR.Printf("Get JsApiTicket from Redis failed,err=%v,token=%s", err, ticket)
		tmpl, _ := template.ParseFiles("./template/minions_record.html")
		tmpl.Execute(resp, nil)
		return
	}
	wxappid, err := config.GetConfigString(config.Key_WeiXinAppId)
	if err != nil {
		l.ERROR.Printf("Get Key_WeiXinAppId failed,err=%v,token=%s", err, ticket)
		tmpl, _ := template.ParseFiles("./template/minions_record.html")
		tmpl.Execute(resp, nil)
		return
	}
	timestamp := time.Now().Unix()
	nonceStr := util.Str2sha1(fmt.Sprintf("%d", time.Now().UnixNano()))
	wxconfig := WxConfig{
		Debug:     true,
		AppId:     wxappid,
		Timestamp: timestamp,
		NonceStr:  nonceStr,
	}
	wxconfig.Signature = util.Str2sha1(fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%d&url=%s", ticket, nonceStr, timestamp, "http://"+req.Request.Host+req.Request.RequestURI))
	wxconfig.JsApiList = []string{"startRecord", "stopRecord", "onRecordEnd", "playVoice", "pauseVoice", "stopVoice", "uploadVoice", "downloadVoice", "onMenuShareTimeline", "onMenuShareAppMessage", "translateVoice"}
	tmpl, err := template.ParseFiles("./template/minions_record.html")
	if err != nil {
		l.ERROR.Printf("ParseFile ./template/minions_record.html failed ,err :=%v ", err)
		return
	}
	err = tmpl.Execute(resp, MinionsData{wxconfig, "", 0})
	if err != nil {
		l.ERROR.Printf("Execute ./template/minions_record.html failed ,err :=%v ", err)
		return
	}
}

type MediaId struct {
	MediaId   string `json:"MediaId"`
	MinionsId int64  `json:"MinionsId"`
}
type MediaUrl struct {
	Url     string `json:"Url"`
	ShareId string `json:"ShareId"`
}

func IsContainUnSafeStr(str string) bool {
	IsContain := false
	IsContain = strings.Contains(str, `.`)
	IsContain = strings.Contains(str, `/`)
	IsContain = strings.Contains(str, `?`)
	IsContain = strings.Contains(str, `%`)
	IsContain = strings.Contains(str, `<`)
	IsContain = strings.Contains(str, `>`)
	IsContain = strings.Contains(str, `&`)
	IsContain = strings.Contains(str, `\`)
	return IsContain
}

func getId(request *restful.Request, response *restful.Response) {
	var req MediaId
	var rsp MediaUrl
	err := request.ReadEntity(&req)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.WriteEntity(&rsp)
		return
	}
	if len(req.MediaId) != 64 {
		InsertErrLog("Invalid MediaId", "Count not Equal 64", req.MediaId, util.GetClientIP(request))
		response.WriteHeader(http.StatusForbidden)
		response.WriteEntity(&rsp)
		return
	}
	if IsContainUnSafeStr(req.MediaId) == true {
		InsertErrLog("Invalid MediaId", "Contain UnSafeStr", req.MediaId, util.GetClientIP(request))
		response.WriteHeader(http.StatusForbidden)
		response.WriteEntity(&rsp)
		return
	}
	ts := time.Now().UnixNano()
	key := fmt.Sprintf(`voice/%d.mp3`, ts)
	fileurl, err := util.ConvertFromWeiXin(req.MediaId)
	if err != nil {
		if err.Error() == "Tooshort" {
			return
		}
		l.ERROR.Println("convert audio from WeiXin Failed! MediaId:=%s,err:=%v", req.MediaId, err)
		response.WriteHeader(http.StatusInternalServerError)
		response.WriteEntity(&rsp)
		return
	}
	err = util.UploadToQiniu(fileurl, key)
	if err != nil {
		l.ERROR.Println("Upload to Qiniu Failed! filepath:=%s,key:=%s,err:=%v", fileurl, key, err)
		response.WriteHeader(http.StatusInternalServerError)
		response.WriteEntity(&rsp)
		return
	}
	rsp.Url = `http://7xlago.com2.z0.glb.qiniucdn.com/` + key
	rsp.ShareId = fmt.Sprintf("%d", ts)
	response.WriteEntity(&rsp)
	err = os.Remove(fileurl)
	if err != nil {
		l.ERROR.Println("delete temp file failed!filepath:=%s!err:=%v", fileurl, err)
	}
	var mediaLog schema.MediaLog
	mediaLog.CreateTs = int(time.Now().Unix())
	mediaLog.Ip = util.GetClientIP(request)
	mediaLog.Key = fmt.Sprintf("%d", ts)
	mediaLog.MediaId = req.MediaId
	mediaLog.RefId = req.MinionsId
	_, err = db.DB.InsertOne(&mediaLog)
	if err != nil {
		l.ERROR.Println("insert medialog into db failed!MediaId:=%s!err:=%v", req.MediaId, err)
	}
}

func InsertErrLog(errType string, errContent string, MediaId string, Ip string) {
	var errLog schema.ErrLog
	errLog.CreateTs = int(time.Now().Unix())
	errLog.ErrContent = errContent
	errLog.ErrType = errType
	errLog.MediaId = MediaId
	errLog.Ip = Ip
	_, err := db.DB.InsertOne(&errLog)
	if err != nil {
		l.ERROR.Println("insert ErrLog into db failed!MediaId:=%s!err:=%v", MediaId, err)
	}
}
