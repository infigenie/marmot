/*
	版权所有，侵权必究
	署名-非商业性使用-禁止演绎 4.0 国际
	警告： 以下的代码版权归属hunterhug，请不要传播或修改代码
	你可以在教育用途下使用该代码，但是禁止公司或个人用于商业用途(在未授权情况下不得用于盈利)
	商业授权请联系邮箱：gdccmcm14@live.com QQ:459527502

	All right reserved
	Attribution-NonCommercial-NoDerivatives 4.0 International
	Notice: The following code's copyright by hunterhug, Please do not spread and modify.
	You can use it for education only but can't make profits for any companies and individuals!
	For more information on commercial licensing please contact hunterhug.
	Ask for commercial licensing please contact Mail:gdccmcm14@live.com Or QQ:459527502
*
*/

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/hunterhug/marmot/miner"
	"github.com/hunterhug/marmot/util"
)

// Open http://music.163.com/#/playlist?id=145258012
const (
	SuggestionUrl = "http://sug.music.baidu.com/info/suggestion"
	Fmlink        = "http://music.baidu.com/data/music/fmlink"
)

var Refer string
var Sp, _ = miner.New(nil)

func init() {
	Sp.SetUa(miner.RandomUa())
	fmt.Println(`
	----------
	地址类似：http://music.163.com/#/playlist?id=145258012
	参考：https://github.com/lifei6671/NeteaseCloudMusicFlac
	----------
	`)
}

func main() {
	M := ""
	for M == "" {
		M = util.Input("输入网易云链接：", "")
	}
	fmt.Println("开始欣赏： " + M)

	nurl := strings.Replace(M, "#/", "", -1)

	Refer = nurl

	// http://music.163.com/#/playlist?id=145258012
	response, err := DownloadString(nurl, nil)
	if err != nil {
		fmt.Println("获取远程URL内容时出错：", err)
		return
	}

	dir, _ := util.GetCurrentPath()

	dir = filepath.Join(dir, "songs_dir")

	err = util.MakeDir(dir)
	if err != nil {
		fmt.Println(err.Error())
	}

	reg := regexp.MustCompile(`<ul class="f-hide">(.*?)</ul>`)

	mm := reg.FindAllString(string(response), -1)

	waitGroup := sync.WaitGroup{}

	if len(mm) > 0 {
		reg = regexp.MustCompile(`<li><a .*?>(.*?)</a></li>`)

		contents := mm[0]
		urlli := reg.FindAllSubmatch([]byte(contents), -1)

		for _, item := range urlli {

			query := url.Values{}
			query.Set("word", string(item[1]))
			query.Set("version", "2")
			query.Set("from", "0")

			res, err := DownloadString(SuggestionUrl, query)
			if err != nil {
				fmt.Println("获取音乐列表时出错：", err)
				continue
			}

			var dat map[string]interface{}

			err = json.Unmarshal([]byte(res), &dat)

			if err != nil {
				fmt.Println("反序列化JSON时出错:", err)
				continue
			}

			if _, ok := dat["data"]; ok == false {
				fmt.Println("没有找到音乐资源:", string(item[1]))
				continue
			}

			songid := dat["data"].(map[string]interface{})["song"].([]interface{})[0].(map[string]interface{})["songid"].(string)

			query = url.Values{}
			query.Set("songIds", songid)
			query.Set("type", "flac")

			res, err = DownloadString(Fmlink, query)

			if err != nil {
				fmt.Println("获取音乐文件时出错：", err)
				continue
			}

			var data map[string]interface{}

			err = json.Unmarshal(res, &data)

			if code, ok := data["errorCode"]; (ok && code.(float64) == 22005) || err != nil {
				fmt.Println("解析音乐文件时出错：", err)
				continue
			}

			songlink := data["data"].(map[string]interface{})["songList"].([]interface{})[0].(map[string]interface{})["songLink"].(string)

			r := []rune(songlink)
			if len(r) < 10 {
				fmt.Println("没有无损音乐地址:", string(item[1]), songlink)
				continue
			} else {
				fmt.Println("存在无损音乐地址:", string(item[1]), songlink)
			}

			songname := data["data"].(map[string]interface{})["songList"].([]interface{})[0].(map[string]interface{})["songName"].(string)

			artistName := data["data"].(map[string]interface{})["songList"].([]interface{})[0].(map[string]interface{})["artistName"].(string)

			filename := filepath.Join(dir, songname+"-"+artistName+".flac")
			filenametemp := filepath.Join(dir, songname+"-"+artistName+".flacxx")
			if util.FileExist(filename) {
				continue
			}
			waitGroup.Add(1)
			go func() {
				fmt.Println("正在下载 ", songname, " ......")
				defer waitGroup.Done()

				songRes, err := http.Get(songlink)
				if err != nil {
					fmt.Println("下载文件时出错：", songlink)
					return
				}

				songFile, err := os.Create(filenametemp)
				written, err := io.Copy(songFile, songRes.Body)
				if err != nil {
					fmt.Println("保存临时音乐文件时出错：", err)
					return
				} else {
					errr := util.Rename(filenametemp, filename)
					if errr != nil {
						fmt.Println("临时文件重命名失败:" + filenametemp)
					}
				}
				fmt.Println(songname, "下载完成,文件大小：", fmt.Sprintf("%.2f", (float64(written) / (1024 * 1024))), "MB")
			}()

		}

	}
	waitGroup.Wait()
}

func DownloadString(remoteUrl string, queryValues url.Values) (body []byte, err error) {
	body = nil
	uri, err := url.Parse(remoteUrl)
	if err != nil {
		return
	}
	if queryValues != nil {
		values := uri.Query()
		if values != nil {
			for k, v := range values {
				queryValues[k] = v
			}
		}
		uri.RawQuery = queryValues.Encode()
	}
	url := uri.String()
	Sp.SetUrl(url)
	Sp.SetRefer(Refer)
	response, err := Sp.Get()
	if err != nil {
		return
	}

	if Sp.Response.StatusCode == 200 {
		switch Sp.Response.Header.Get("Content-Encoding") {
		case "gzip":
			reader, _ := gzip.NewReader(bytes.NewReader(response))
			for {
				buf := make([]byte, 1024)
				n, err := reader.Read(buf)

				if err != nil && err != io.EOF {
					panic(err)
				}

				if n == 0 {
					break
				}
				body = append(body, buf...)
			}
		default:
			return response, nil

		}
	}
	return
}
