package platform

import (
	"context"
	"net/http"
	"time"

	"github.com/beck-8/subs-check/config"
)

func CheckAlive(ctx context.Context, httpClient *http.Client) (bool, error) {
	// 如果上层没有超时控制，这里保证最小超时
	if _, ok := ctx.Deadline(); !ok {
		timeout := time.Duration(config.GlobalConfig.Timeout) * time.Millisecond
		if timeout < 5*time.Second {
			timeout = 5 * time.Second
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.GlobalConfig.AliveTestUrl, nil)
	if err != nil {
		return false, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	// 2xx
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, nil
	}
	return false, nil
}
