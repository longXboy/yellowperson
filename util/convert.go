package util

import (
	"flag"
	"fmt"
	"github.com/krig/go-sox"
	"io/ioutil"
	"net/http"
	"sync"
	"yellowPerson/cache"
	l "yellowPerson/util/log"
)

var Lock *sync.Mutex

const (
	MAX_SAMPLES = 2048
)

func flow(in, out *sox.Format, samples []sox.Sample) {
	n := uint(len(samples))
	for number_read := in.Read(samples, n); number_read > 0; number_read = in.Read(samples, n) {
		out.Write(samples, uint(number_read))
	}
}

func ConvertFromWeiXin(mediaId string) (string, error) {
	var samples [MAX_SAMPLES]sox.Sample

	flag.Parse()

	/*content, err := ioutil.ReadFile("haha.amr")
	if err != nil {
		log.Fatal("cant open file!err:=%v", err)
	}*/
	token, err := cache.GetToken()
	if err != nil {
		l.ERROR.Println("can't get wx token from redis,err:=%v", err)
		return "", err
	}
	downloadurl := fmt.Sprintf(`http://file.api.weixin.qq.com/cgi-bin/media/get?access_token=%s&media_id=%s`, token, mediaId)
	content, err := httpdownload(downloadurl)
	if err != nil {
		l.ERROR.Println("download audio from WeiXin failed!,err:=%v", err)
		return "", err
	}
	if string(content[:7]) != string([]byte{35, 33, 65, 77, 82, 10, 12}) {
		l.ERROR.Println("invalid content!,content:=%v", content)
		return "", err
	}
	if len(content) < 400 {
		return "", fmt.Errorf("Tooshort")
	}
	Lock.Lock()
	defer Lock.Unlock()
	in := sox.OpenMemRead(content)
	if in == nil {
		l.ERROR.Println("Failed to open memory Read !,err:=%v", err)
		return "", err
	}
	// Set up the memory buffer for writing
	buf := sox.NewMemstream()
	defer buf.Release()
	out := sox.OpenMemstreamWrite(buf, in.Signal(), nil, "sox")
	if out == nil {
		l.ERROR.Println("Failed to open memory buffer!,err:=%v", err)
		in.Release()
		return "", err
	}
	flow(in, out, samples[:])
	out.Release()
	in.Release()

	in = sox.OpenMemRead(buf)
	if in == nil {
		l.ERROR.Println("Failed to open memory buffer !,err:=%v", err)
		return "", err
	}
	filepath := fmt.Sprintf(`./tempfiles/%s.mp3`, mediaId)
	out = sox.OpenWrite(filepath, in.Signal(), nil, "mp3")
	if out == nil {
		l.ERROR.Println("Failed to open memory Read !,err:=%v", err)
		in.Release()
		return "", err
	}
	chain := sox.CreateEffectsChain(in.Encoding(), out.Encoding())
	// Make sure to clean up!
	var e *sox.Effect

	// The first effect in the effect chain must be something that can
	// source samples; in this case, we use the built-in handler that
	// inputs data from an audio file.
	e = sox.CreateEffect(sox.FindEffect("input"))
	e.Options(in)
	// This becomes the first "effect" in the chain
	chain.Add(e, in.Signal(), in.Signal())
	e.Release()

	e = sox.CreateEffect(sox.FindEffect("bend"))
	e.Options("0,700,0.1")
	chain.Add(e, in.Signal(), in.Signal())
	e.Release()

	/*e = sox.CreateEffect(sox.FindEffect("pitch"))
	e.Options("-250")
	chain.Add(e, in.Signal(), in.Signal())
	e.Release()*/
	// The last effect in the effect chain must be something that only consumes
	// samples; in this case, we use the built-in handler that outputs data.
	e = sox.CreateEffect(sox.FindEffect("output"))
	e.Options(out)
	chain.Add(e, in.Signal(), in.Signal())
	e.Release()

	// Flow samples through the effects processing chain until EOF is reached.
	chain.Flow()
	//flowandwrite(in, out, samples[:])

	out.Release()
	in.Release()

	return filepath, nil
}

func httpdownload(uri string) ([]byte, error) {
	res, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("ReadFile: Size of download: %d\n", len(d))
	return d, err
}
