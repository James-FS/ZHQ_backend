package utils

import (
	"encoding/json"

	"fmt"
	"net/http"
	"zhq-backend/config"
)

// WeChatLoginResponse 微信登录响应结构
type WeChatLoginResponse struct {
	OpenID     string `json:"openid"`      //用户的唯一标识符
	SessionKey string `json:"session_key"` //会话密钥，用于后续的加密解密
	UnionID    string `json:"unionid"`     //用户在微信开放平台的唯一标识符
	ErrCode    int    `json:"errcode"`     //错误代码
	ErrMsg     string `json:"errmsg"`      //错误信息
}

// GetWeChatOpenID 调用微信接口获取openid和session_key
func GetWeChatOpenID(code string) (*WeChatLoginResponse, error) {
	appID := config.GetString("wechat.appid")
	secret := config.GetString("wechat.secret")

	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, secret, code)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //确保资源被正确释放，即使中间发生错误或提前返回

	//JSON解码
	var result WeChatLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信接口错误：%s", result.ErrMsg)
	}

	return &result, nil
}
