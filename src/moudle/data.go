package moudle

import "encoding/xml"

//视频信息
type VideoList struct {
	//视频id
	Aid          string `bson:"aid"`
	//作者id
	Mid        string  `bson:"mid"`
	Author       string `bson:"authorid"`
	Coins        string `bson:"coins"`
	Pic          string `bson:"pic"`
	Title        string `bson:"title"`
	//视频弹幕列表id
	Cid          string `bson:"cid"`
	//播放量
	Play         string `bson:"play"`
	//得分
	Pts          string `bson:"pts"`
	//弹幕量
	Video_review string `bson:"video_review"`
	//分类id
	Rid          string `bson:"rid"`
	TimeStamp    string `bson:"timeStamp"`
	//分类
	Rank         int    `bson:"rank"`
	DanMuCheck   bool   `bson:"danMuCheck"`
}

//视频信息
type UpperInfo struct {
	//作者id
	Mid        string  `bson:"mid"`
	Name       string  `bson:"name"`
	Pic        string  `bson:"pic"`
	DanMuCheck bool    `bson:"danMuCheck"`
	VideoDanMu  []DanMu `bson:"videoDanMu"`
}

//单条弹幕信息
type DanMu struct {

	//视频id
	Cid       string `bson:"cid"`
	//发送者id
	Uid       string `bson:"uid"`
	TimeStamp string `bson:"timeStamp"`
	//弹幕id
	Did       string `bson:"did"`
	Content   string `bson:"content"`
}

//一个视频的所有弹幕
type VideoDanMu struct {
	//弹幕集id
	Cid       string  `bson:"cid"`
	//upper主id
	Mid       string `bson:"mid"`
	DanMuList []DanMu `bson:"danmuList"`
}

//解析弹幕xml用到的xml
type Recurlyservers struct {
	XMLName xml.Name `xml:"i"`
	DS      []D      `xml:"d"`
}
type D struct {
	P       string `xml:"p,attr"`
	Content string `xml:",chardata"`
}
