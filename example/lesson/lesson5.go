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

/*
	Proxy  Worker!
	You first should own a remote machine, Then in your local tap:
		`ssh -ND 1080 ubuntu@remoteIp`
	It will gengerate socks5 proxy client in your local, which port is 1080
*/

package main

import (
	"fmt"
	"os"

	"github.com/hunterhug/marmot/expert"
	"github.com/hunterhug/marmot/miner"
)

func init() {
	miner.SetLogLevel(miner.DEBUG)
}

func main() {
	// You can use a lot of proxy ip such "https/http/socks5"
	proxy_ip := "socks5://127.0.0.1:1080"

	url := "https://www.google.com"

	worker, err := miner.New(proxy_ip)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	body, err := worker.SetUa(miner.RandomUa()).SetUrl(url).SetMethod(miner.GET).Go()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(parse(body))
	}
}

// Parse HTML page
func parse(data []byte) string {
	doc, err := expert.QueryBytes(data)
	if err != nil {
		fmt.Println(err.Error())
	}
	return doc.Find("title").Text()
}
