package proxies

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"log/slog"

	"github.com/beck-8/subs-check/config"
	"github.com/metacubex/mihomo/common/convert"
)

func geoTimeout() time.Duration {
	t := time.Duration(config.GlobalConfig.Timeout) * time.Millisecond
	if t <= 0 {
		t = 3 * time.Second
	}
	if t > 8*time.Second {
		t = 8 * time.Second
	}
	return t
}

func newGeoRequest(method, url, ua string) (*http.Request, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), geoTimeout())
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	if ua != "" {
		req.Header.Set("User-Agent", ua)
	}
	return req, cancel, nil
}

func GetProxyCountry(httpClient *http.Client) (loc string, ip string) {
	for i := 0; i < config.GlobalConfig.SubUrlsReTry; i++ {
		// 优先使用更稳定的 ipapi.co
		loc, ip = GetIPAPI(httpClient)
		if loc != "" && ip != "" {
			return
		}
		// 次选 ip-api.com，速率宽松
		loc, ip = GetIPAPICom(httpClient)
		if loc != "" && ip != "" {
			return
		}
	}
	return
}

func GetIPAPICom(httpClient *http.Client) (loc string, ip string) {
	type GeoIPData struct {
		Query       string `json:"query"`
		CountryCode string `json:"countryCode"`
		Status      string `json:"status"`
		Message     string `json:"message"`
	}

	req, cancel, err := newGeoRequest(http.MethodGet, "http://ip-api.com/json/?fields=status,message,countryCode,query", convert.RandUserAgent())
	if err != nil {
		slog.Debug(fmt.Sprintf("创建请求失败: %s", err))
		return
	}
	defer cancel()

	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Debug(fmt.Sprintf("ip-api获取节点位置失败: %s", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug(fmt.Sprintf("ip-api返回非200状态码: %v", resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Debug(fmt.Sprintf("ip-api读取节点位置失败: %s", err))
		return
	}

	var geo GeoIPData
	err = json.Unmarshal(body, &geo)
	if err != nil {
		slog.Debug(fmt.Sprintf("解析ip-api JSON 失败: %v", err))
		return
	}

	if geo.Status != "success" {
		slog.Debug(fmt.Sprintf("ip-api返回状态非success: %s", geo.Message))
		return
	}

	return geo.CountryCode, geo.Query
}

func GetIPAPI(httpClient *http.Client) (loc string, ip string) {
	type GeoIPData struct {
		IP      string `json:"ip"`
		Country string `json:"country_code"`
	}

	// ipapi.co 免费接口，限制较宽松
	req, cancel, err := newGeoRequest(http.MethodGet, "https://ipapi.co/json", convert.RandUserAgent())
	if err != nil {
		slog.Debug(fmt.Sprintf("创建请求失败: %s", err))
		return
	}
	defer cancel()

	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Debug(fmt.Sprintf("ipapi获取节点位置失败: %s", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug(fmt.Sprintf("ipapi返回非200状态码: %v", resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Debug(fmt.Sprintf("ipapi读取节点位置失败: %s", err))
		return
	}

	var geo GeoIPData
	err = json.Unmarshal(body, &geo)
	if err != nil {
		slog.Debug(fmt.Sprintf("解析ipapi JSON 失败: %v", err))
		return
	}

	return geo.Country, geo.IP
}

func GetEdgeOneProxy(httpClient *http.Client) (loc string, ip string) {
	type GeoResponse struct {
		Eo struct {
			Geo struct {
				CountryCodeAlpha2 string `json:"countryCodeAlpha2"`
			} `json:"geo"`
			ClientIp string `json:"clientIp"`
		} `json:"eo"`
	}

	url := "https://functions-geolocation.edgeone.app/geo"
	req, cancel, err := newGeoRequest(http.MethodGet, url, convert.RandUserAgent())
	if err != nil {
		slog.Debug(fmt.Sprintf("创建请求失败: %s", err))
		return
	}
	defer cancel()

	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Debug(fmt.Sprintf("edgeone获取节点位置失败: %s", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug(fmt.Sprintf("edgeone返回非200状态码: %v", resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Debug(fmt.Sprintf("edgeone读取节点位置失败: %s", err))
		return
	}

	var eo GeoResponse
	err = json.Unmarshal(body, &eo)
	if err != nil {
		slog.Debug(fmt.Sprintf("解析edgeone JSON 失败: %v", err))
		return
	}

	return eo.Eo.Geo.CountryCodeAlpha2, eo.Eo.ClientIp
}

func GetCFProxy(httpClient *http.Client) (loc string, ip string) {
	url := "https://www.cloudflare.com/cdn-cgi/trace"
	req, cancel, err := newGeoRequest(http.MethodGet, url, convert.RandUserAgent())
	if err != nil {
		slog.Debug(fmt.Sprintf("创建请求失败: %s", err))
		return
	}
	defer cancel()

	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Debug(fmt.Sprintf("cf获取节点位置失败: %s", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug(fmt.Sprintf("cf返回非200状态码: %v", resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Debug(fmt.Sprintf("cf读取节点位置失败: %s", err))
		return
	}

	// Parse the response text to find loc=XX
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "loc=") {
			loc = strings.TrimPrefix(line, "loc=")
		}
		if strings.HasPrefix(line, "ip=") {
			ip = strings.TrimPrefix(line, "ip=")
		}
	}
	return
}

func GetIPLark(httpClient *http.Client) (loc string, ip string) {
	type GeoIPData struct {
		IP      string `json:"ip"`
		Country string `json:"country_code"`
	}

	url := string([]byte{104, 116, 116, 112, 115, 58, 47, 47, 102, 51, 98, 99, 97, 48, 101, 50, 56, 101, 54, 98, 46, 97, 97, 112, 113, 46, 110, 101, 116, 47, 105, 112, 97, 112, 105, 47, 105, 112, 99, 97, 116})
	req, cancel, err := newGeoRequest(http.MethodGet, url, "curl/8.7.1")
	if err != nil {
		slog.Debug(fmt.Sprintf("创建请求失败: %s", err))
		return
	}
	defer cancel()

	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Debug(fmt.Sprintf("iplark获取节点位置失败: %s", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug(fmt.Sprintf("iplark返回非200状态码: %v", resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Debug(fmt.Sprintf("iplark读取节点位置失败: %s", err))
		return
	}

	var geo GeoIPData
	err = json.Unmarshal(body, &geo)
	if err != nil {
		slog.Debug(fmt.Sprintf("解析iplark JSON 失败: %v", err))
		return
	}

	return geo.Country, geo.IP
}

func GetMe(httpClient *http.Client) (loc string, ip string) {
	type GeoIPData struct {
		IP      string `json:"ip"`
		Country string `json:"country_code"`
	}

	url := "https://ip.122911.xyz/api/ipinfo"
	req, cancel, err := newGeoRequest(http.MethodGet, url, "subs-check (https://github.com/beck-8/subs-check)")
	if err != nil {
		slog.Debug(fmt.Sprintf("创建请求失败: %s", err))
		return
	}
	defer cancel()

	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Debug(fmt.Sprintf("me获取节点位置失败: %s", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug(fmt.Sprintf("me返回非200状态码: %v", resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Debug(fmt.Sprintf("me读取节点位置失败: %s", err))
		return
	}

	var geo GeoIPData
	err = json.Unmarshal(body, &geo)
	if err != nil {
		slog.Debug(fmt.Sprintf("解析me JSON 失败: %v", err))
		return
	}

	return geo.Country, geo.IP
}
