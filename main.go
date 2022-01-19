package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ringo199/live_forward/model"
)

func getRequest(apiUrl string, params *map[string]string, header *map[string]string) (*http.Request, error) {
	data := url.Values{}
	for k, v := range *params {
		data.Set(k, v)
	}
	u, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		panic(err)
	}
	u.RawQuery = data.Encode() // URL encode
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		panic(err)
	}
	for k, v := range *header {
		req.Header.Add(k, v)
	}
	return req, err
}

func getRoomPlayInfo(cid string) *map[string][]string {
	client := &http.Client{}
	params := map[string]string{
		"room_id":    cid,
		"no_playurl": "0",
		"mask":       "1",
		"qn":         "0",
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
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var res model.Resp

	json.Unmarshal(body, &res)

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
	return &_map
}

func getASoulStream(w http.ResponseWriter, r *http.Request) {
	ASoul_cid := map[string]string{
		"ava":    "22625025",
		"bella":  "22632424",
		"carol":  "22634198",
		"diana":  "22637261",
		"eileen": "22625027",
	}
	urlInfo := make(map[string][]string)
	for _, cid := range ASoul_cid {
		tmpInfo := *getRoomPlayInfo(cid)
		if tmpInfo["http_hls"] != nil {
			urlInfo = tmpInfo
		}
	}
	if urlInfo["http_hls"] != nil {
		http.Redirect(w, r, urlInfo["http_hls"][0], http.StatusMovedPermanently)
	} else {
		fmt.Fprintln(w, "直播间未开启😭")
	}
}

func getUrl(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	cid := r.Form.Get("cid")
	urlInfo := *getRoomPlayInfo(cid)
	if urlInfo["http_hls"] != nil || urlInfo["http_stream"] != nil {
		urlInfoJson, _ := json.Marshal(urlInfo)
		fmt.Fprintln(w, string(urlInfoJson))
	} else {
		fmt.Fprintln(w, "直播间未开启")
	}
}

func getStream(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	cid := r.Form.Get("cid")
	urlInfo := *getRoomPlayInfo(cid)

	if urlInfo["http_hls"] != nil {
		http.Redirect(w, r, urlInfo["http_hls"][0], http.StatusMovedPermanently)
	} else if urlInfo["http_stream"] != nil {
		http.Redirect(w, r, urlInfo["http_stream"][0], http.StatusMovedPermanently)
	} else {
		fmt.Fprintln(w, "直播间未开启")
	}
}

func main() {
	http.HandleFunc("/", getASoulStream)
	http.HandleFunc("/get", getStream)
	http.HandleFunc("/getUrl", getUrl)

	err := http.ListenAndServe(":7732", nil)
	if err != nil {
		fmt.Printf("http server failed, err:%v\n", err)
		return
	}
}
