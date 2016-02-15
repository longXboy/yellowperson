package config

import (
	conf "github.com/robfig/config"
	slog "log"
)

const Key_WeiXinAppId = "wx.appid"
const Key_WeiXinAppSecret = "wx.appsecret"

const Key_RedisAddr = "redis.addr"
const Key_RedisPsw = "redis.psw"
const Key_RedisDb = "redis.db"

const Key_DbUser = "db.user"
const Key_DbPsw = "db.psw"
const Key_Dbname = "db.name"
const Key_Addr = "db.addr"

var defaultConfig *conf.Config
var env map[string]string

func GetConfigString(key string) (v string, err error) {
	v, err = defaultConfig.String(env["sec"], key)
	return
}

func GetConfigInt(key string) (v int, err error) {
	v, err = defaultConfig.Int(env["sec"], key)
	return
}

func GetConfigInt64(key string) (v int64, err error) {
	v32, er := defaultConfig.Int(env["sec"], key)
	v = int64(v32)
	err = er
	return
}

func GetConfigBool(key string) (v bool, err error) {
	v, err = defaultConfig.Bool(env["sec"], key)
	return
}

func GetConfigFloat(key string) (v float64, err error) {
	v, err = defaultConfig.Float(env["sec"], key)
	return
}

func LoadConfigFile() error {
	configpath, _ := env["config"]
	c, err := conf.ReadDefault(configpath)
	if err != nil {
		return err
	}
	defaultConfig.Merge(c)
	return nil
}

func InitConfigEnv(en map[string]string) {
	var err error
	env = en
	_, ok := env["config"]
	_, ok1 := env["sec"]
	if !ok || !ok1 {
		slog.Fatalf("config and section not found in Env\n")
	}
	configpath, _ := env["config"]
	defaultConfig, err = conf.ReadDefault(configpath)
	if err != nil {
		slog.Fatalf("ReadDefault failed. err=%v\n", err)
	}
}
