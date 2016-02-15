package util

import (
	"github.com/qiniu/api.v6/auth/digest"
	"github.com/qiniu/api.v6/conf"
	fio "github.com/qiniu/api.v6/io"
	"github.com/qiniu/api.v6/rs"
)

func UploadToQiniu(path string, key string) error {
	mac := digest.Mac{
		"9tU_z0oBbZqlPTOhn4JCE0zG9-1OKQemVC8FhCMn",
		[]byte("KuRYsC7CekI5BEFNxFs65lB3buW7LvpmhP9X98hA"),
	}
	policy := rs.PutPolicy{}
	policy.Scope = "yellowperson"
	putExtra := fio.PutExtra{}
	putExtra.MimeType = "audio/mpeg"
	conf.UP_HOST = "http://up.qiniug.com"
	uptoken := policy.Token(&mac)
	putRet := fio.PutRet{}
	/*fStat, err := os.Stat(path)
	if err != nil {
		return err
	}*/
	//	fsize := fStat.Size()
	//	startTime := time.Now()
	err := fio.PutFile(nil, &putRet, uptoken, key, path, &putExtra)
	if err != nil {
		return err
	}
	//lastNano := time.Now().UnixNano() - startTime.UnixNano()
	//lastTime := fmt.Sprintf("%.2f", float32(lastNano)/1e9)
	//avgSpeed := fmt.Sprintf("%.1f", float32(fsize)*1e6/float32(lastNano))
	//fmt.Println("newVersion", "Last time:", lastTime, "s, Average Speed:", avgSpeed, "KB/s")
	return nil
}
