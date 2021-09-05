package bot

import (
	"base/pkg/helper"
	"base/pkg/library/line/bot/template"
	"base/pkg/model"
	"base/pkg/model/bot"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	underscore "github.com/ahl5esoft/golang-underscore"
	"github.com/line/line-bot-sdk-go/linebot"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func GroupParseTextGenTemplate(lineId LineID, text string) interface{} {
	parseText := strings.Split(text, "|")

	//功能說明
	if len(parseText) == 1 {
		switch parseText[0] {
		case "c helper", "記帳說明", "記帳":
			return linebot.NewTextMessage("*記帳*\n將按照群組人數去做平均計算，使用記帳請使用以下格式輸入\n\"記帳|標題|總金額|備註\"\n例如:\n記帳|生日聚餐|1234|本人生日")
		case "c list helper", "記帳列表說明":
			return linebot.NewTextMessage("*記帳列表*\n將回傳記帳紀錄的列表，格式為:\n日期時間 標題|金額| 平均金額 |付款人|備註")
		case "c balance helper", "記帳結算說明", "結算說明":
			return linebot.NewTextMessage("*記帳結算說明*\n將刪除記帳資料，格式為:\n記帳結算|日期(可選)")
		}
	}
	//功能
	switch parseText[0] {
	// c list||記帳列表
	case "c list", "記帳列表":
		messages := []linebot.SendingMessage{}

		lb := LineBilling{}
		where := make(map[string]interface{})
		where["group_id"] = lineId.GroupID
		not := make(map[string]interface{})
		lbs, err := lb.Get(where, not)
		if err != nil {
			log.Fatalf("Get failed: %v", err)
		}
		//沒有記帳資料
		if len(lbs) == 0 {
			return linebot.NewTextMessage("目前沒有記帳紀錄哦！")
		}
		//user id line member display name
		dstByUserID := make(map[string]string, 0)
		//user id total amount
		lbUserIDAmount := make(map[string]float64, 0)
		underscore.Chain(lbs).DistinctBy("UserID").SelectMany(func(lb LineBilling, _ int) map[string]string {
			dst := make(map[string]string)
			lineMember, err := botClient.GetGroupMemberProfile(lb.GroupID, lb.UserID).Do()
			if err != nil {
				dst[lb.UserID] = "Unkonw"
				return dst
			}
			dst[lb.UserID] = lineMember.DisplayName
			return dst
		}).Value(&dstByUserID)
		var sbList strings.Builder
		sbList.Grow(len(lbs))
		for _, lb := range lbs {
			var memberName string
			amountAvg, amountAvgBase := calculateAmount(lineId.GroupID, helper.ConvertToFloat64(lb.Billing.Amount))
			//check line member display name is exist
			if _, ok := dstByUserID[lb.UserID]; ok {
				memberName = dstByUserID[lb.UserID]
				lbUserIDAmount[lb.UserID] = lbUserIDAmount[lb.UserID] + amountAvg
			}
			text := lb.Billing.CreatedAt.Format(helper.Yyyymmddhhmmss) + " " +
				lb.Billing.Title + "|" + helper.ConvertToString(lb.Billing.Amount) + "/" + helper.ConvertToString(amountAvgBase) + " = " + helper.ConvertToString(amountAvg) + " |" + memberName + "|" + lb.Billing.Note + "\n"
			sbList.WriteString(text)
		}
		messages = append(messages, linebot.NewTextMessage(sbList.String()))
		//billing list string
		var sbTotal strings.Builder
		sbTotal.Grow(len(dstByUserID))
		text := "總付款金額：\n"
		sbTotal.WriteString(text)
		for userID, name := range dstByUserID {
			text = fmt.Sprintf("%v: *%v*\n", name, helper.ConvertToString(lbUserIDAmount[userID]))
			sbTotal.WriteString(text)
		}
		messages = append(messages, linebot.NewTextMessage(sbTotal.String()))

		return messages
	// c||記帳|生日聚餐|1234|本人生日
	case "c", "記帳":
		title := parseText[1]
		amount := helper.ConvertToInt(parseText[2])
		note := ""
		//如果有輸入備註
		if len(parseText) == 4 {
			note = parseText[3]
		}
		billingAction(lineId, amount, title, note)
		amountFloat64 := helper.ConvertToFloat64(amount)
		amountAvg, amountAvgBase := calculateAmount(lineId.GroupID, amountFloat64)
		return linebot.NewTextMessage(title + ":記帳完成," + parseText[2] + "/" + helper.ConvertToString(int(amountAvgBase)) + " = " + "*" + helper.ConvertToString(amountAvg) + "*")
	// 記帳結算
	case "記帳結算", "結算":
		date := time.Now().Format(helper.Iso8601)
		//如果有輸入限制日期
		if len(parseText) == 2 {
			date = parseText[1]
		}
		var lbs []bot.LineBilling
		err := model.DB.Preload("Billing").Where("updated_at < ?", date).Find(&lbs).Error
		if err != nil {
			log.Fatalf("Get failed: %v", err)
		}
		//沒有記帳資料
		if len(lbs) == 0 {
			return linebot.NewTextMessage(fmt.Sprintf("%v以前沒有記帳紀錄哦！", date))
		}
		listText := getLineBillingList(lineId, lbs)
		//template
		postBack := LinePostBackAction{
			Action: "結算",
			Data: map[string]interface{}{
				"Date": date,
			},
		}
		postBackJson, err := json.Marshal(postBack)
		if err != nil {
			log.Fatalf("Marshal failed: %v", err)
		}
		leftBtn := linebot.NewPostbackAction("是", string(postBackJson), "", "")
		rightBtn := linebot.NewMessageAction("否", "記帳列表")
		confirmTemplate := linebot.NewConfirmTemplate(listText, leftBtn, rightBtn)
		return linebot.NewTemplateMessage("確定要刪除這些紀錄?", confirmTemplate)
	case "我的大頭貼":
		lineMember, err := botClient.GetGroupMemberProfile(lineId.GroupID, lineId.UserID).Do()
		if err != nil {
			return nil
		}
		return linebot.NewImageMessage(lineMember.PictureURL, lineMember.PictureURL)
	case "TODO":
		date := parseText[1]
		replyText := parseText[2]
		parseDate := strings.Split(date, " ")
		switch parseDate[0] {
		case "every":
			// TODO|every 19:55|測試29號13:30送出
			todoAction(lineId.UserID, "every", parseDate[1], template.TODO(replyText))
			return linebot.NewTextMessage("設置完成將於每天" + parseDate[1] + "\n傳送訊息:" + replyText)
		default:
			// TODO|2020/02/29 13:00|測試29號13:30送出
			todoAction(lineId.UserID, "once", date, template.TODO(replyText))
			return linebot.NewTextMessage("設置完成將於" + date + "\n傳送訊息:" + replyText)
		}

	}
	if helper.ConvertToBool(viper.Get("LINE_MESSAGING_DEBUG").(string)) {
		return linebot.NewTextMessage("目前沒有此功能")
	}
	return nil
}

func GroupParsePostBackGenTemplate(lineId LineID, postBack *linebot.Postback) interface{} {
	data := []byte(postBack.Data)
	lpba := LinePostBackAction{}
	err := json.Unmarshal(data, &lpba)
	if err != nil {
		log.Fatalf("line group GroupParsePostBackGenTemplate json unmarshal error: %v", err)
	}
	switch lpba.Action {
	case "結算":
		date := lpba.Data["Date"].(string)
		var lbs []bot.LineBilling
		err := model.DB.Where("updated_at < ?", date).Delete(&lbs).Error
		if err != nil {
			log.Fatalf("line group GroupParsePostBackGenTemplate 結算 Delete failed: %v", err)
		}
		return linebot.NewTextMessage(fmt.Sprintf("成功刪除*%v*以前的記帳資料", date))
	}
	return nil
}

func getLineBillingList(lineId LineID, lbs []bot.LineBilling) string {
	//user id line member display name
	dstByUserID := make(map[string]string, 0)
	underscore.Chain(lbs).DistinctBy("UserID").SelectMany(func(lb bot.LineBilling, _ int) map[string]string {
		dst := make(map[string]string)
		lineMember, err := botClient.GetGroupMemberProfile(lb.GroupID, lb.UserID).Do()
		if err != nil {
			dst[lb.UserID] = "Unkonw"
			return dst
		}
		dst[lb.UserID] = lineMember.DisplayName
		return dst
	}).Value(&dstByUserID)
	var sbList strings.Builder
	sbList.Grow(len(lbs))
	for _, lb := range lbs {
		var memberName string
		amountAvg, amountAvgBase := calculateAmount(lineId.GroupID, helper.ConvertToFloat64(lb.Billing.Amount))
		//check line member display name is exist
		if _, ok := dstByUserID[lb.UserID]; ok {
			memberName = dstByUserID[lb.UserID]
		}
		text := lb.Billing.CreatedAt.Format(helper.Yyyymmddhhmmss) + " " +
			lb.Billing.Title + "|" + helper.ConvertToString(lb.Billing.Amount) + "/" + helper.ConvertToString(amountAvgBase) + " = " + helper.ConvertToString(amountAvg) + " |" + memberName + "|" + lb.Billing.Note + "\n"
		sbList.WriteString(text)
	}
	return string(sbList.String())
}

func calculateAmount(groupID string, amount float64) (float64, int) {
	//預設平均計算基數
	amountAvgBase := 3.0
	groupMemberCount, err := botClient.GetGroupMemberCount(groupID).Do()
	if err != nil {
		log.Fatal("line messaging api get group member count error:", err)
	}
	amountAvgBase = helper.ConvertToFloat64(groupMemberCount.Count)
	amountAvg := amount / amountAvgBase
	return amountAvg, groupMemberCount.Count
}