package cache

import (
	"gopkg.in/redis.v3"
	slog "log"
	"time"
	"yellowPerson/util/config"
	"yellowPerson/util/log"
	wx "yellowPerson/wxInterface"
)

var redisclient *redis.Client

func InitRedis() {
	redisaddr, err := config.GetConfigString(config.Key_RedisAddr)
	if err != nil {
		slog.Fatalf("GetConfigString redis.addr from  failed.\n")
		return
	}
	redispasswd, err := config.GetConfigString(config.Key_RedisPsw)
	if err != nil {
		slog.Fatalf("GetConfigString redis.psw from configfailed.\n")
		return
	}
	redisdb, err := config.GetConfigInt(config.Key_RedisDb)
	if err != nil {
		slog.Fatalf("GetConfigString redis.db from config failed.\n")
		return
	}
	redisclient = redis.NewClient(&redis.Options{
		Addr:         redisaddr,
		Password:     redispasswd,    // no password set
		DB:           int64(redisdb), // use default DB
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
		PoolSize:     64,
	})
	_, err = redisclient.Ping().Result()
	if err != nil {
		slog.Fatalln("redis client ping failed. " + err.Error())
	}
}

func Close() {
	redisclient.Close()
}

const (
	Cache_WeiXinAccToken    = "yellowperson.wx_acctoken"
	Cache_WeiXinjsApiTicket = "yellowperson.wx_jsapiticket"
)

func GetToken() (string, error) {
	accToken, err := redisclient.Get(Cache_WeiXinAccToken).Result()
	if err != nil && err != redis.Nil {
		log.ERROR.Printf("Get wx_acctoken failed from Redis!err=%v", err)
		return "", err
	}
	if err == redis.Nil {
		wxtoken, err := wx.GetWeiXinToken()
		if err != nil {
			log.ERROR.Printf("Get wx_acctoken failed via HTTP!err=%v", err)
			return "", err
		}
		_, err = redisclient.Set(Cache_WeiXinAccToken, wxtoken.Token, time.Duration((wxtoken.ExpireIn-5)*1000000000)).Result()
		if err != nil {
			log.ERROR.Printf("Set wx_acctoken failed from Redis!err=%v", err)
			return "", err
		}
		return wxtoken.Token, nil
	}
	return accToken, nil
}

func GetJsApiTicket() (string, error) {
	ticket, err := redisclient.Get(Cache_WeiXinjsApiTicket).Result()
	if err != nil && err != redis.Nil {
		log.ERROR.Printf("Get wx_jsapiticket failed from Redis!err=%v", err)
		return "", err
	}
	if err == redis.Nil {
		token, err := GetToken()
		if err != nil || token == "" {
			log.ERROR.Printf("Get Token from Redis failed,err=%v,token=%s", err, token)
			return "", nil
		}
		wxticket, err := wx.GetWeiXinJsApi(token)
		if err != nil {
			log.ERROR.Printf("Get wx_JsApiTicket failed via HTTP!err=%v", err)
			return "", err
		}
		_, err = redisclient.Set(Cache_WeiXinjsApiTicket, wxticket.Ticket, time.Duration((wxticket.ExpireIn-5)*1000000000)).Result()
		if err != nil {
			log.ERROR.Printf("Set wx_jsapiticket failed from Redis!err=%v", err)
			return "", err
		}
		return wxticket.Ticket, nil

	}
	return ticket, nil
}
