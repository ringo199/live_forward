package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ringo199/live_forward/model"
)

func quality2qn(quality string) string {
	if quality == "" {
		quality = "999"
	}
	tmp, _ := strconv.Atoi(quality)
	if tmp > 5 {
		tmp = 5
	}
	switch tmp {
	case 0:
		return "0"
	case 1:
		return "80"
	case 2:
		return "150"
	case 3:
		return "400"
	case 4:
		return "10000"
	case 5:
		return "20000"
	default:
		return "20000"
	}
}

func map2Onlyurl(info map[string][]string) string {
	if info == nil {
		return ""
	}
	rand.Seed(time.Now().UnixNano())
	if info["http_hls"] != nil {
		i := rand.Intn(len(info["http_hls"]))
		if len(info["http_hls"]) == 1 && i == 0 {
			i++
		}
		return info["http_hls"][i]
	} else if info["http_stream"] != nil {
		return info["http_stream"][rand.Intn(len(info["http_stream"]))]
	} else {
		return ""
	}
}

func getRequest(apiUrl string, params *map[string]string, header *map[string]string) (*http.Request, error) {
	data := url.Values{}
	if params != nil {
		for k, v := range *params {
			data.Set(k, v)
		}
	}
	u, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		return nil, err
	}
	u.RawQuery = data.Encode() // URL encode
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if header != nil {
		for k, v := range *header {
			req.Header.Add(k, v)
		}
	}
	return req, nil
}

func getRoomPlayInfo(cid string, quality string) (*map[string][]string, error) {
	client := &http.Client{}
	qn := quality2qn(quality)
	params := map[string]string{
		"room_id":    cid,
		"no_playurl": "0",
		"mask":       "1",
		"qn":         qn,
		"platform":   "web",
		"protocol":   "0,1",
		"format":     "0,2",
		"codec":      "0,1",
	}
	header := map[string]string{
		"referrer":   "https://live.bilibili.com/",
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.104 Safari/537.36",
	}
	apiUrl := "https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo"
	req, _ := getRequest(apiUrl, &params, &header)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res model.Resp

	json.Unmarshal(body, &res)

	// if len(res.Data.Playurl_info.Playurl.Stream) == 0 {
	// 	return nil, fmt.Errorf("直播间未开启")
	// }
	var _map map[string][]string = make(map[string][]string)
	for _, stream := range res.Data.Playurl_info.Playurl.Stream {
		_map[stream.Protocol_name] = make([]string, 0)
		for _, form := range stream.Format {
			for _, coc := range form.Codec {
				for _, url_info := range coc.Url_info {
					_url := url_info.Host + coc.Base_url + url_info.Extra
					_map[stream.Protocol_name] = append(_map[stream.Protocol_name], _url)
				}
			}
		}
	}
	return &_map, nil
}

func getVideoRealUrl(BVCode string) (string, error) {
	client := &http.Client{}
	header := map[string]string{
		"referrer":   "https://www.bilibili.com/",
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.104 Safari/537.36",
	}
	apiUrl := "https://www.bilibili.com/video/" + BVCode
	req, _ := getRequest(apiUrl, nil, &header)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// doc, err := goquery.NewDocumentFromReader(resp.Body)
	// if err != nil {
	// 	return "", err
	// }
	// script := doc.Find("script")

	return string(body), nil
}

func getASoulStream(w http.ResponseWriter, r *http.Request) {
	ASoul_cid := map[string]string{
		"ava":    "22625025",
		"bella":  "22632424",
		"carol":  "22634198",
		"diana":  "22637261",
		"eileen": "22625027",
	}
	r.ParseForm()
	qn := r.Form.Get("qn")
	urlInfo := make(map[string][]string)
	for _, cid := range ASoul_cid {
		tmpInfo, err := getRoomPlayInfo(cid, qn)
		if err == nil {
			urlInfo = *tmpInfo
			break
		}
	}
	url := map2Onlyurl(urlInfo)
	if url != "" {
		http.Redirect(w, r, url, http.StatusMovedPermanently)
	} else {
		// todo: 转发本地视频流
		fmt.Fprintln(w, "直播间未开启😭")
	}
}

func getUrl(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	cid := r.Form.Get("cid")
	qn := r.Form.Get("qn")
	urlInfo, _ := getRoomPlayInfo(cid, qn)

	url := map2Onlyurl(*urlInfo)
	if url != "" {
		urlInfoJson, _ := json.Marshal(urlInfo)
		fmt.Fprintln(w, string(urlInfoJson))
	} else {
		fmt.Fprintln(w, "直播间未开启")
	}
}

func getStream(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	cid := r.Form.Get("cid")
	qn := r.Form.Get("qn")
	urlInfo, _ := getRoomPlayInfo(cid, qn)

	url := map2Onlyurl(*urlInfo)
	if url != "" {
		http.Redirect(w, r, url, http.StatusMovedPermanently)
	} else {
		fmt.Fprintln(w, "直播间未开启")
	}
}

func getVideoUrl(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	BVCode := r.Form.Get("BV")
	body, err := getVideoRealUrl(BVCode)
	if err != nil {
		fmt.Fprintln(w, err)
	} else {
		fmt.Fprintln(w, body)
	}
}

func main() {
	http.HandleFunc("/", getASoulStream)
	http.HandleFunc("/get", getStream)
	http.HandleFunc("/getUrl", getUrl)
	http.HandleFunc("/getVideoUrl", getVideoUrl)

	err := http.ListenAndServe(":7732", nil)
	if err != nil {
		fmt.Printf("http server failed, err:%v\n", err)
		return
	}

}
