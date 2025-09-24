package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/noah-loop/backend/modules/notify/internal/application/service"
	"github.com/noah-loop/backend/modules/notify/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// AliyunSMSProvider 阿里云短信提供商
type AliyunSMSProvider struct {
	logger infrastructure.Logger
	client *http.Client
}

// NewAliyunSMSProvider 创建阿里云短信提供商
func NewAliyunSMSProvider(logger infrastructure.Logger) service.SMSProvider {
	return &AliyunSMSProvider{
		logger: logger,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendSMS 发送短信
func (p *AliyunSMSProvider) SendSMS(ctx context.Context, data *service.SMSData, config *domain.ChannelConfig) error {
	p.logger.Info("Sending SMS via Aliyun",
		zap.String("phone", data.Phone),
		zap.String("content", data.Content))

	// 获取配置
	accessKey, _ := config.GetConfig("access_key")
	secretKey, _ := config.GetConfig("secret_key")
	signName, _ := config.GetConfig("sign_name")
	templateCode, _ := config.GetConfig("template_code")
	region, _ := config.GetConfig("region")

	if region == "" {
		region = "cn-hangzhou"
	}

	// 构建请求参数
	params := map[string]string{
		"AccessKeyId":      accessKey,
		"Action":           "SendSms",
		"Format":           "JSON",
		"PhoneNumbers":     data.Phone,
		"SignName":         signName,
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"SignatureMethod":  "HMAC-SHA256",
		"SignatureVersion": "1.0",
		"SignatureNonce":   strconv.FormatInt(time.Now().UnixNano(), 10),
		"Version":          "2017-05-25",
	}

	// 使用模板或直接内容
	if data.TemplateID != "" || templateCode != "" {
		if data.TemplateID != "" {
			params["TemplateCode"] = data.TemplateID
		} else {
			params["TemplateCode"] = templateCode
		}
		
		// 模板参数
		if data.Variables != nil {
			templateParams, _ := json.Marshal(data.Variables)
			params["TemplateParam"] = string(templateParams)
		}
	}

	// 生成签名
	signature := p.generateSignature(params, secretKey)
	params["Signature"] = signature

	// 构建请求URL
	endpoint := fmt.Sprintf("https://dysmsapi.%s.aliyuncs.com/", region)
	
	// 发送HTTP请求
	resp, err := p.sendHTTPRequest(ctx, endpoint, params)
	if err != nil {
		p.logger.Error("Failed to send HTTP request", zap.Error(err))
		return err
	}

	// 解析响应
	var result AliyunSMSResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查结果
	if result.Code != "OK" {
		return fmt.Errorf("SMS sending failed: code=%s, message=%s", result.Code, result.Message)
	}

	p.logger.Info("SMS sent successfully via Aliyun",
		zap.String("phone", data.Phone),
		zap.String("biz_id", result.BizId))

	return nil
}

// generateSignature 生成签名
func (p *AliyunSMSProvider) generateSignature(params map[string]string, secretKey string) string {
	// 1. 排序参数
	var keys []string
	for k := range params {
		if k != "Signature" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 2. 构建查询字符串
	var query []string
	for _, k := range keys {
		query = append(query, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(params[k])))
	}
	queryString := strings.Join(query, "&")

	// 3. 构建待签名字符串
	stringToSign := "GET&" + url.QueryEscape("/") + "&" + url.QueryEscape(queryString)

	// 4. 计算签名
	key := secretKey + "&"
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature
}

// sendHTTPRequest 发送HTTP请求
func (p *AliyunSMSProvider) sendHTTPRequest(ctx context.Context, endpoint string, params map[string]string) ([]byte, error) {
	// 构建URL参数
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	// 创建请求
	url := endpoint + "?" + values.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// ValidateConfig 验证配置
func (p *AliyunSMSProvider) ValidateConfig(config *domain.ChannelConfig) error {
	requiredFields := []string{"access_key", "secret_key", "sign_name"}
	
	for _, field := range requiredFields {
		if _, exists := config.GetConfig(field); !exists {
			return domain.NewDomainError("MISSING_CONFIG", "missing required Aliyun SMS config: "+field)
		}
	}
	
	return nil
}

// GetProviderName 获取提供商名称
func (p *AliyunSMSProvider) GetProviderName() string {
	return "aliyun"
}

// AliyunSMSResponse 阿里云短信响应
type AliyunSMSResponse struct {
	Message   string `json:"Message"`
	RequestId string `json:"RequestId"`
	BizId     string `json:"BizId"`
	Code      string `json:"Code"`
}
