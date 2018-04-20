package handlers

import (
	"fmt"
	"github.com/devfeel/dotweb"
	"github.com/devfeel/longweb/config"
	. "github.com/devfeel/longweb/const"
	"github.com/devfeel/longweb/framework/log"
	. "github.com/devfeel/longweb/message"
	"strconv"
	"github.com/devfeel/longweb/framework/json"
)


type GroupIDs struct{
	IDs []string
}


//websocket统一处理入口
func OnWebSocket(ctx dotweb.Context) error {


	logTitle := "[OnWebSocket][" + ctx.Request().Url() + "] "

	appId := ctx.QueryString("appid")
	groupId := ctx.QueryString("groupid")
	groupIds := ctx.QueryString("groupids")
	userId := ctx.QueryString("userid")
	from := ctx.QueryString("from")
	token := ctx.QueryString("token")

	logger.Debug(logTitle+"connect [RemoteIp:"+ctx.RemoteIP()+"]", LogTarget_HttpRequest)

	if appId == "" {
		ctx.WebSocket().SendMessage("no permission connect! => appid is empty")
		logger.Warn(logTitle+"no permission connect! => appid is empty", LogTarget_HttpRequest)
		return nil
	}

	if groupId == "" && groupIds == "" {
		ctx.WebSocket().SendMessage("no permission connect! => groupId and groupIds is empty")
		logger.Warn(logTitle+"no permission connect! => groupId and groupIds is empty", LogTarget_HttpRequest)
		return nil
	}

	groupIDs := new(GroupIDs)
	if groupId == ""{
		err := jsonutil.Unmarshal(groupIds, groupIDs)
		if err != nil{
			ctx.WebSocket().SendMessage("no permission connect! => groupIds is not correct format => " + groupIds)
			logger.Warn(logTitle+"no permission connect! => groupIds is not correct format => " + groupIds, LogTarget_HttpRequest)
			return nil
		}

		if len(groupIDs.IDs) <= 0{
			ctx.WebSocket().SendMessage("no permission connect! => groupIds is not contain ids")
			logger.Warn(logTitle+"no permission connect! => groupIds is not contain ids", LogTarget_HttpRequest)
			return nil
		}
	}

	app, exists := config.GetAppInfo(appId)
	if !exists {
		ctx.WebSocket().SendMessage("no permission connect! => appid[" + appId + "] no exists")
		logger.Warn(logTitle+"no permission connect! => appid["+appId+"] no exists", LogTarget_HttpRequest)
		return nil
	}

	//鉴权
	if token != "" {
		retCode, retMsg := CheckAuthToken(app, appId, groupId, groupIds, userId, token)
		if retCode != 0 {
			logger.Warn(logTitle+"CheckAuthToken failed => "+strconv.Itoa(retCode)+":"+retMsg, LogTarget_HttpRequest)
			ctx.WebSocket().SendMessage(retMsg)
			return nil
		}
	}
	isAuth := false
	if token != "" {
		isAuth = true
	}

	client := NewClient(appId, userId, groupId, from, groupIDs.IDs, isAuth, ctx.WebSocket(), nil)
	defer RemoveClient(client)

	if client == nil {
		ctx.WebSocket().SendMessage("no permission connect! => create client failed")
		logger.Warn(logTitle+"no permission connect! => create client failed", LogTarget_HttpRequest)
		return nil
	}

	//注册客户端
	regCode := RegisterClient(client)
	if regCode != 0 {
		client.SendMessage("no permission connect! =>  register client failed " + strconv.Itoa(regCode))
		logger.Warn(logTitle+"["+fmt.Sprint(client)+"] no permission connect! =>  register client failed "+strconv.Itoa(regCode), LogTarget_HttpRequest)
		return nil
	}

	var strMsg string
	var err error
	for {
		if strMsg, err = client.GetWebSocket().ReadMessage(); err != nil {
			client.SendMessage("connect has a error => " + err.Error())
			logger.Warn(logTitle+"["+client.GetClientInfo()+"] ReadMessage error => "+err.Error(), LogTarget_HttpRequest)
			break
		} else {
			//TODO:与业务方接口交互
			strMsg = strMsg

			var isResponse bool
			var responseMsg string
			if isResponse {
				client.SendMessage(responseMsg)
			}
			//TODO:test code,need remove
			//strLog := "[" + time.Now().Format("2006-01-02 15:04:05") + "] Read [" + strClient + "] From [" + client.RemoteAddr() + ":" + strconv.FormatUint(client.GetIndex(), 10) + "]"
			//userGroup.SendMessage(strLog)
			//fmt.Println(strLog)
		}
		//TODO:log the request
	}
	return nil
}