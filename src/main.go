package main

import (
	"compress/flate"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/axgle/mahonia"
	"github.com/mran/billibilidanmu/src/connon"
	"github.com/mran/billibilidanmu/src/moudle"
	"github.com/parnurzeal/gorequest"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

func getAllHotVideoList() {
	var yingshiUrl = "https://api.bilibili.com/x/web-interface/ranking?day=7&type=1&arc_type=0&jsonp=json&rid=%s"
	timeNow := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	for index, rid := range videoURls {
		println(index)
		_, body, _ := requestaa.Get(fmt.Sprintf(yingshiUrl, rid)).End()
		var f interface{}
		d1 := []byte(body)
		_ = json.Unmarshal(d1, &f)
		m := f.(map[string]interface{})
		vdata := m["data"].(map[string]interface{})
		vList := vdata["list"].([]interface{})

		for index, value := range vList {
			value := value.(map[string]interface{})
			vItem := moudle.VideoList{Aid: fmt.Sprintf("%s", value["aid"]),
				Author:       fmt.Sprint(value["author"]),
				Coins:        fmt.Sprintf("%v", value["coins"]),
				Pic:          fmt.Sprintf("%s", value["pic"]),
				Mid:          fmt.Sprintf("%s", value["mid"]),
				Title:        fmt.Sprintf("%v", value["title"]),
				Cid:          fmt.Sprintf("%.0f", value["cid"]),
				Play:         fmt.Sprintf("%.0f", value["play"]),
				Pts:          fmt.Sprintf("%.0f", value["pts"]),
				Video_review: fmt.Sprintf("%.0f", value["video_review"]),
				Rid:          fmt.Sprintf("%v", rid),
				TimeStamp:    fmt.Sprintf("%s", timeNow),
				Rank:         index,
				DanMuCheck:   false,
			}
			if _, err := connon.VideoCollection.InsertOne(connon.DbContext, vItem); err != nil {
				fmt.Println(err)
			}
			fmt.Println(index)
		}
		fmt.Println(index)
	}
}
func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

/*
根据视频获得弹幕列表
*/
func getDanMuListByVideoList() {
	timeNow := time.Now().AddDate(0, 0, -10).Format("2006-01-02")

	fs, _ := connon.VideoCollection.Find(connon.DbContext, bson.D{{"timeStamp", timeNow}, {"danMuCheck", false}})
	var allUnCheckedVideo []moudle.VideoList
	_ = fs.All(connon.DbContext, &allUnCheckedVideo)
	for index, v := range allUnCheckedVideo {

		fmt.Print(time.Now().Format("01-02 15:04:05"), index, v.Title)

		_, er := getDanMuDetail(v.Mid, v.Cid)
		if er == nil {
			connon.VideoCollection.FindOneAndUpdate(connon.DbContext, bson.D{{"cid", v.Cid}}, bson.D{{"$set", bson.D{{"danMuCheck", true}}}})
		}
		time.Sleep(2 * time.Second)

	}
}
//获取弹幕集的弹幕
func getDanMuDetail(mid string, cid string) (*moudle.VideoDanMu, error) {
	var videoDanMu moudle.VideoDanMu
	videoDanMu.Cid = cid
	videoDanMu.Mid = mid
	lDMs := videoDanMu.DanMuList
	mDML := map[string]moudle.DanMu{}
	for day := 0; day < 1; day++ {
		//timeNow := time.Now().AddDate(0, 0, -day).Format("2006-01-02")
		//或得历史弹幕
		//urlTemplate := fmt.Sprintf("https://api.bilibili.com/x/v2/dm/history?type=1&oid=%s&date=%s", cid, timeNow)
		//直取第一页
		urlTemplate := fmt.Sprintf("https://api.bilibili.com/x/v1/dm/list.so?oid=%s", cid)
		re11, body, er1 := requestaa.Get(urlTemplate).
			AppendHeader("Cookie", "LIVE_BUVID=AUTO9815392643967286; sid=92j9docd; stardustvideo=1; buvid3=2A764511-2EFB-4BC5-A851-9D0178AC765C163046infoc; CURRENT_FNVAL=16; fts=1539866668; CURRENT_QUALITY=80; UM_distinctid=16a0fdb59658fc-081ecff037fda7-e323069-1fa400-16a0fdb5966a4d; rpdid=|(u|u)um|Rk)0J'ullYuuum~|; DedeUserID=576062; DedeUserID__ckMd5=de90b86a5b8f1202; SESSDATA=2774b97a%2C1566969339%2Ceecdb071; bili_jct=89350f9f90d5a88447bfcffb6d05bcdb; bp_t_offset_576062=288147839766090050").
			End()
		re11.Header.Get("123")
		if er1 != nil {
			fmt.Printf("er1: %v", er1)
			return nil, er1[0]
			
		}
		v := moudle.Recurlyservers{}
		//flate格式的解码
		fw := flate.NewReader(strings.NewReader(body))
		var trBody string

		if b, err := ioutil.ReadAll(fw); err == nil {
			trBody = string(b)
		} else {
			fmt.Printf("%v", err)
			fmt.Printf("%v", body)

			return nil, err
		}

		err := xml.Unmarshal([]byte(trBody), &v)
		if err != nil {
			fmt.Printf("xml.Unmarshalerror: %v", err)
			return nil, err
		}
		for _, i := range v.DS {
			psl := strings.Split(i.P, ",")
			ditem := moudle.DanMu{Cid: cid, Uid: psl[3], TimeStamp: psl[4], Did: psl[7], Content: i.Content}
			mDML[ditem.Did] = ditem
		}
	}
	for _, v := range mDML {
		lDMs = append(lDMs, v)
	}
	videoDanMu.DanMuList = lDMs
	_, er2 := connon.DanmuCollection.InsertOne(connon.DbContext, videoDanMu)
	if er2 != nil {
		fmt.Printf("DanmuCollection %v", er2)
		return nil, er2
	}
	fmt.Println("done")
	return &videoDanMu, nil
}

//获取一个upper的所有视频的弹幕
func getDanMuListByUpper() {

	for _, uperId := range uppers {
		var sUperInfo moudle.UpperInfo
		re := connon.UpperCollection.FindOne(connon.DbContext, bson.D{{"mid", uperId}})
		reerr := re.Decode(&sUperInfo)
		//如果存在,//不必重新写入数据库了
		if reerr == nil {
			if sUperInfo.DanMuCheck == true{
				continue
			}
		}
		time.Sleep(2 * time.Second)

		info := getUpperInfo(uperId)
		if info.DanMuCheck {
			continue
		}
		videoLis, _ := getUpperAllVideo(uperId)
		for index, video := range videoLis {
			if index > 1000 {
				break
			}
			re := connon.VideoCollection.FindOne(connon.DbContext, bson.D{{"aid", video.Aid}})
			var reVideo moudle.VideoList
			err := re.Decode(&reVideo)
			if err == nil {
				continue
			}

			fmt.Print(video.Title)
			chatID := getChatId(video.Aid)
			if chatID == nil {
				continue
			}
			video.Cid = *chatID
			_, _ = connon.VideoCollection.InsertOne(connon.DbContext, video)

			danmu, err := getDanMuDetail(uperId, *chatID)
			if err == nil {
				_, _ = connon.DanmuCollection.InsertOne(connon.DbContext, danmu)
				connon.VideoCollection.FindOneAndUpdate(connon.DbContext, bson.D{{"aid", video.Aid}}, bson.D{{"$set", bson.D{{"danMuCheck", true}}}})
			}
			time.Sleep(1 * time.Second)

		}
		connon.UpperCollection.FindOneAndUpdate(connon.DbContext, bson.D{{"mid", uperId}}, bson.D{
			{"$set", bson.D{
				{"danMuCheck", true},
			}}})
	}
}

/*
aid,视频id
返回视频的弹幕集chatid
*/
func getChatId(aid string) *string {
	urlTemplate := fmt.Sprintf("https://www.bilibili.com/video/av%s", aid)
	_, body, er1 := requestaa.Get(urlTemplate).
		Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.100 Safari/537.36").
		End()
	if (er1 != nil) {
		fmt.Println(er1)
		return nil
	}
	if (strings.Contains(body, "视频不见了，你可以试试")) {
		return nil
	}

	flysnowRegexp := regexp.MustCompile("=[0-9]*&aid")
	params := flysnowRegexp.FindStringSubmatch(body)

	for _, param := range params {
		chatid := fmt.Sprintf(param[1:strings.Index(param, "&aid")])
		return &chatid
	}
	return nil

}
//获得UP主的信息
func getUpperInfo(mid string) *moudle.UpperInfo {
	//获取upper的信息
	upperInfoUrl := fmt.Sprintf("https://api.bilibili.com/x/space/acc/info?mid=%s&jsonp=jsonp", mid)
	_, body, er1 := requestaa.Get(upperInfoUrl).
		Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.100 Safari/537.36").
		End()
	if (er1 != nil) {
		fmt.Println(er1)
		return nil
	}
	var f interface{}
	dBody := []byte(body)
	_ = json.Unmarshal(dBody, &f)
	jBody := f.(map[string]interface{})
	jData := jBody["data"].(map[string]interface{})
	upInfo := moudle.UpperInfo{Mid: fmt.Sprintf("%.0f", jData["mid"]),
		Name:       fmt.Sprintf("%s", jData["name"]),
		Pic:        fmt.Sprintf("%s", jData["face"]),
		DanMuCheck: false,
	}
	var sUperInfo moudle.UpperInfo
	re := connon.UpperCollection.FindOne(connon.DbContext, bson.D{{"mid", mid}})
	reerr := re.Decode(&sUperInfo)
	//如果存在,//不必重新写入数据库了
	if reerr == nil {
		upInfo.DanMuCheck = true
		return &upInfo
	}
	_, _ = connon.UpperCollection.InsertOne(connon.DbContext, upInfo)
	return &upInfo
}
//获得up主的所有视频
func getUpperAllVideo(mid string) ([]moudle.VideoList, error) {
	pagesCount := 1
	videoLists := []moudle.VideoList{}
	timeNow := time.Now().Format("2006-01-02")

	//分页获取
	for index := 1; index <= pagesCount; index++ {
		//获取upper的信息
		upperInfoUrl := fmt.Sprintf("https://space.bilibili.com/ajax/member/getSubmitVideos?mid=%s&pagesize=30&tid=0&page=%v&keyword=&order=pubdate", mid, index)
		_, body, er1 := requestaa.Get(upperInfoUrl).
			Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.100 Safari/537.36").
			//AppendHeader("Cookie", "LIVE_BUVID=AUTO9815392643967286; sid=92j9docd; stardustvideo=1; buvid3=2A764511-2EFB-4BC5-A851-9D0178AC765C163046infoc; CURRENT_FNVAL=16; fts=1539866668; CURRENT_QUALITY=80; UM_distinctid=16a0fdb59658fc-081ecff037fda7-e323069-1fa400-16a0fdb5966a4d; rpdid=|(u|u)um|Rk)0J'ullYuuum~|; DedeUserID=576062; DedeUserID__ckMd5=de90b86a5b8f1202; SESSDATA=2774b97a%2C1566969339%2Ceecdb071; bili_jct=89350f9f90d5a88447bfcffb6d05bcdb; bp_t_offset_576062=288147839766090050").
			End()
		if (er1 != nil) {
			fmt.Println(er1)
			return nil, er1[0]
		}
		var f interface{}
		dBody := []byte(body)
		_ = json.Unmarshal(dBody, &f)
		jBody := f.(map[string]interface{})
		jData := jBody["data"].(map[string]interface{})
		pagesCount = int(jData["pages"].(float64))
		jvList := jData["vlist"].([]interface{})
		for _, vItem := range jvList {
			vvItem := vItem.(map[string]interface{})
			//fmt.Println(vvItem["title"])
			vItem := moudle.VideoList{Aid: fmt.Sprintf("%.0f", vvItem["aid"]),
				Author:       fmt.Sprint(vvItem["author"]),
				Pic:          fmt.Sprintf("%s", vvItem["pic"]),
				Mid:          fmt.Sprintf("%.0f", vvItem["mid"]),
				Title:        fmt.Sprintf("%v", vvItem["title"]),
				Play:         fmt.Sprintf("%.0f", vvItem["play"]),
				Video_review: fmt.Sprintf("%.0f", vvItem["video_review"]),
				Rid:          fmt.Sprintf("%0.f", vvItem["typeid"]),
				TimeStamp:    fmt.Sprintf("%s", timeNow),
				Rank:         index,
				DanMuCheck:   false,
			}
			videoLists = append(videoLists, vItem)
		}
	}
	return videoLists, nil
}

var requestaa *gorequest.SuperAgent
var videoURls []string
var uppers []string

func main() {
	videoURls = []string{
		//connon.AremenRid,
		connon.AdonghuaRid,
		connon.AguochuangRid,
		connon.AyinyueRid,
		connon.AwudaoRid,
		connon.AyouxiRid,
		connon.AkejiRid,
		connon.AshumaRid,
		connon.AshenghuoRid,
		connon.AguichuRid,
		connon.AshishangRid,
		connon.AyuleRid,
		connon.AyingshiRid,
	}
	uppers = []string{
		"546195",
		"777536",
		"122879",
		"9824766",
		"1532165",
		"883968",
		"20165629",
		"375375",
		"4162287",
		"32786875",
		"321173469",
		"250858633",
		"196356191",
		"1577804",
		"562197",
		"466272",
		//"221648",
		"176037767",
		"168598",
		"927587",
		"15773384",
		"17409016",
		"1565155",
		"16794231",
		"1935882",
		"33683045",
		"14110780",
		"13354765",
		"62540916",
		"585267",
		"808171",
		"7487399",
		"116683",
		"290526283",
		"1643718",
		"8578857",
		"9008159",
		"32820037",
		"39847479",
		"433351",
		"8960728",
		"50329118",
		"486183",
		"10558098",
		"2403047",
		"322892",
		"7584632",
		"2920960",
		"4474705",
		"123938419",
		"174501086",
		"63231",
		"391679",
		"6574487",
		"423895",
		"163637592",
		"280793434",
		"10462362",
		"43536",
		"52250",
		"52250",
		"17819768",
		"19577966",
		"390461123",
		"161775300",
		"51896064",
		"3766866",
		"21837784",
		"282994",
		"289887832",
		"1328260",
		"2206456",
		"113362335",
		"26366366",
		"79061224",
		"730732",
		"398510",
		"8366990",
		"8047632",
		"268104",
		"27218150",
		"9550310",
		"104207471",
		"1740850",
		"10330740",
		"161419374",
		"3682229",
		"19642758",
		"234256",
		"7875104",
		"43389575",
		"27534330",
		"250111460",
		"30643878",
		"15967711",
		"295711424",
		"617285",
		"99157282",
		"164139557",
		"7714",
	}
	requestaa = gorequest.New()
	//获得所有分区的视频列表
	//getAllHotVideoList()
	fmt.Println(time.Now())
	//getDanMuListByVideoList()
	/*cisd, _ := getUpperAllVideo("3379951")
	println(len(cisd))*/
	//getChatId("55287468")
	getDanMuListByUpper()
	fmt.Println(time.Now())
	//_ = getDanMuDetail("108485733")
}
