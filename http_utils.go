// @author      Liu Yongshuai<liuyongshuai@hotmail.com>
// @date        2018-11-22 18:37

package goUtils

import "strings"

//对原始cookie进行五马分尸
func SplitRawCookie(ck string) (ret map[string]string) {
	ret = make(map[string]string)
	ck = strings.TrimSpace(ck)
	if len(ck) == 0 {
		return
	}
	kvs := strings.Split(ck, ";")
	if len(kvs) == 0 {
		return
	}
	for _, val := range kvs {
		val = strings.TrimSpace(val)
		if !strings.Contains(val, "=") {
			continue
		}
		ind := strings.Index(val, "=")
		k := strings.TrimSpace(val[0:ind])
		v := strings.TrimSpace(val[ind+1:])
		if len(k) == 0 || len(v) == 0 {
			continue
		}
		ret[k] = v
	}
	return
}

//合并cookie
func JoinRawCookie(ck map[string]string) (ret string) {
	if ck == nil {
		return ""
	}
	if len(ck) == 0 {
		return ""
	}
	var tmp []string
	for k, v := range ck {
		tmp = append(tmp, k+"="+v)
	}
	ret = strings.Join(tmp, "; ")
	return
}
