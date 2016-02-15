package util

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"sort"
	"strconv"
	"time"
	"yellowPerson/util/log"
)

type MsgEncryptFrom struct {
	ToUserName string `xml:"ToUserName"`
	Encrypt    string `xml:"Encrypt"`
}

type MsgRawFrom struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int      `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
	MsgId        int64    `xml:"MsgId"`
}

type MsgRawTo struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int      `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
}

type MsgEncryptTo struct {
	XMLName      xml.Name `xml:"xml"`
	MsgSignature string   `xml:"MsgSignature"`
	TimeStamp    int      `xml:"TimeStamp"`
	Nonce        string   `xml:"Nonce"`
	Encrypt      string   `xml:"Encrypt"`
}

func WxCheckValidate(timestamp string, nonce string, data []byte, msg_sig string) bool {
	ts, err := strconv.ParseInt(timestamp, 10, 32)
	if err != nil {
		log.ERROR.Printf("Time Stamp Invalied!ts:=%s", timestamp)
		return false
	}
	if (int(time.Now().Unix()) - int(ts)) > 45 {
		log.ERROR.Printf("Time Stamp Expired!now:=%d,ts:=%d", time.Now().Unix(), ts)
		return false
	}
	v := MsgEncryptFrom{}
	err = xml.Unmarshal(data, &v)
	if err != nil {
		log.ERROR.Printf("Xml(%s) Unmarshal enrypted data failed!err:=%v", string(data), err)
		return false
	}
	token := "mkwhatyellowpersonlongxia"
	tmps := []string{token, timestamp, nonce, v.Encrypt}
	sort.Strings(tmps)
	Signature := Str2sha1(tmps[0] + tmps[1] + tmps[2] + tmps[3])
	if msg_sig == Signature {
		return true
	} else {
		return false
	}
}

func WxXmlDecode(data []byte) (raw MsgRawFrom, err error) {
	v := MsgEncryptFrom{}
	err = xml.Unmarshal(data, &v)
	if err != nil {
		log.ERROR.Printf("Xml(%s) Unmarshal enrypted data failed!err:=%v", string(data), err)
		return
	}
	if len(v.Encrypt) < 40 {
		log.ERROR.Printf("v.Encrypt data(%s) is too short!err:=%v", string(v.Encrypt))
		err = fmt.Errorf("v.Encrypt data too short")
		return
	}
	aesMsg, _ := base64.StdEncoding.DecodeString(v.Encrypt)
	aesKey, _ := base64.StdEncoding.DecodeString("RQNdXk4AxP1RMOfLisAoE1Uo1BrlO7dbP14H1EREqvl=")
	content, err := AesDecrypt(aesMsg, aesKey)
	if err != nil {
		log.ERROR.Printf("AesDecrypt(%s) failed!err:=%v", string(aesMsg), err)
		return
	}
	if len(content) < 40 {
		log.ERROR.Printf("raw data(%s) is too short!err:=%v", string(content))
		err = fmt.Errorf("raw data too short")
		return
	}
	if string(content[len(content)-18:]) != "wx45502c38a5b98c56" {
		log.ERROR.Printf("AppId(%s) not compared!err:=%v", string(content[len(content)-18:]), err)
		err = fmt.Errorf("AppId not match")
		return
	}
	b_buf := bytes.NewBuffer(content[16:20])
	var lens int32
	err = binary.Read(b_buf, binary.BigEndian, &lens)
	if err != nil {
		log.ERROR.Printf("Read raw data lens(%s) failed!err:=%v", string(content[16:20]), err)
		return
	}
	if int(lens) != len(content[20:len(content)-18]) {
		log.ERROR.Printf("AppId(%s) not compared!err:=%v", string(content[len(content)-18:]), err)
		err = fmt.Errorf("lens not match")
		return
	}
	err = xml.Unmarshal(content[20:len(content)-18], &raw)
	if err != nil {
		log.ERROR.Printf("Xml(%s) Unmarshal raw data failed!err:=%v", string(content[20:len(content)-18]), err)
		return
	}
	if (int(time.Now().Unix()) - raw.CreateTime) > 45 {
		log.ERROR.Printf("Time stamp Expired!now:=%v,ts:=%d", int(time.Now().Unix()), raw.CreateTime)
		return
	}
	return
}

func WxXmlEncode(raw MsgRawTo) (data []byte, err error) {
	//xmlRaw, _ := xml.Marshal(raw)
	str := "<xml><ToUserName><![CDATA[%s]]></ToUserName><FromUserName><![CDATA[%s]]></FromUserName><CreateTime>%d</CreateTime><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[%s]]></Content></xml>"
	//fmt.Println(string(xmlRaw))
	xmlRaw := fmt.Sprintf(str, raw.ToUserName, raw.FromUserName, raw.CreateTime, raw.Content)
	fmt.Println(xmlRaw)
	lens := len(xmlRaw)
	msgLen := bytes.NewBuffer([]byte{})
	binary.Write(msgLen, binary.BigEndian, int32(lens))
	random := time.Now().UnixNano()
	randomRaw := fmt.Sprintf("%d", random)
	msgRandom := Str2sha1(randomRaw)[:16]
	charRaw := make([]byte, 0)
	charRaw = append(charRaw, []byte(msgRandom)...)
	charRaw = append(charRaw, msgLen.Bytes()...)
	charRaw = append(charRaw, []byte(xmlRaw)...)
	charRaw = append(charRaw, []byte("wx45502c38a5b98c56")...)
	//	strRaw := msgRandom + string(msgLen.Bytes()) + xmlRaw + "wx45502c38a5b98c56"
	aesKey, _ := base64.StdEncoding.DecodeString("RQNdXk4AxP1RMOfLisAoE1Uo1BrlO7dbP14H1EREqvl=")
	encrypted, err := AesEncrypt(charRaw, aesKey)
	if err != nil {
		log.ERROR.Printf("AesEncrypt(%s) failed!err:=%v", string(charRaw), err)
		return
	}
	ts := int(time.Now().Unix())
	nonce := time.Now().UnixNano()
	v := MsgEncryptTo{}
	encrypted64 := base64.StdEncoding.EncodeToString(encrypted)
	v.Encrypt = encrypted64
	token := "mkwhatyellowpersonlongxia"
	tmps := []string{token, fmt.Sprintf("%d", ts), fmt.Sprintf("%d", nonce), encrypted64}
	sort.Strings(tmps)
	v.MsgSignature = Str2sha1(tmps[0] + tmps[1] + tmps[2] + tmps[3])
	v.Nonce = fmt.Sprintf("%d", nonce)
	v.TimeStamp = int(ts)
	//v.XMLName = xml.Name{Local: "xml", Space: ""}
	//data, _ = xml.Marshal(v)
	str = "<xml><Encrypt><![CDATA[%s]]></Encrypt><MsgSignature><![CDATA[%s]]></MsgSignature><TimeStamp>%d</TimeStamp><Nonce><![CDATA[%s]]></Nonce></xml>"
	data = []byte(fmt.Sprintf(str, v.Encrypt, v.MsgSignature, v.TimeStamp, v.Nonce))
	return
}
