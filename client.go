package httpretry

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

var defaultRetryStatusMap = map[int]string{
	http.StatusInternalServerError: "서버 처리 불가로 재시도",
	http.StatusBadGateway:          "게이트웨이 오류로 재시도",
	http.StatusServiceUnavailable:  "서비스 사용 불가상태로 재시도",
	http.StatusGatewayTimeout:      "게이트웨이 타임아웃으로 재시도",
}

type retriableTransport struct {
	http.RoundTripper
	requestTimeout   time.Duration
	maxRetries       int
	retryStatusCodes map[int]string
	backoffPolicy    func(attempt int) time.Duration
	debugMode        bool
}

// NewClient HTTP 클라이언트를 생성하고 재시도 설정을 적용
func NewClient(settings *Settings, retryStatusCodes ...int) *http.Client {
	return &http.Client{Transport: newRetriableTransport(settings, retryStatusCodes...)}
}

// newRetriableTransport는 재시도 가능한 Transport를 생성합니다.
func newRetriableTransport(
	settings *Settings,
	retryStatusCodes ...int,
) (customTransport *retriableTransport) {
	if settings == nil {
		settings = NewHTTPSettings()
	}
	var (
		transport *http.Transport
	)
	{
		// transport 설정
		transport = http.DefaultTransport.(*http.Transport)
		transport.MaxIdleConns = settings.MaxIdleConns
		transport.IdleConnTimeout = settings.IdleConnTimeout
		transport.TLSHandshakeTimeout = settings.TLSHandshakeTimeout
		transport.ExpectContinueTimeout = settings.ExpectContinueTimeout
		transport.ResponseHeaderTimeout = settings.ResponseHeaderTimeout
		transport.TLSClientConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		}

		if settings.Insecure {
			transport.TLSClientConfig.InsecureSkipVerify = true
		}
	}
	{
		// customTransport 설정
		retryMap := extendDefault(retryStatusCodes)
		if settings.BackoffPolicy == nil {
			settings.BackoffPolicy = defaultBackoffPolicy
		}
		customTransport = &retriableTransport{
			RoundTripper:     transport,
			requestTimeout:   settings.RequestTimeout,
			maxRetries:       settings.MaxRetry,
			retryStatusCodes: retryMap,
			backoffPolicy:    settings.BackoffPolicy,
			debugMode:        settings.DebugMode,
		}
	}
	return
}

// RoundTrip 요청을 처리하며 재시도를 구현
//
// 각 Request는 Timeout을 가지며, Timeout을 초과하는 경우 재시도를 수행
//
//   - 요청 실패 시, 재시도를 수행
//   - 요청 성공 시, 응답을 반환
//   - 재시도 횟수를 초과하면 에러 반환
func (rt *retriableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var allErrors error // 모든 시도에서 발생한 에러를 저장

	for attempt := 1; attempt <= rt.maxRetries+1; attempt++ {
		// 부모 context가 이미 만료되었는지 확인
		if req.Context().Err() != nil {
			allErrors = multierr.Append(allErrors, errors.New("cancelled from parent context"))
			break
		}

		// 최대 재시도 횟수를 초과하면 종료
		if attempt > rt.maxRetries {
			allErrors = multierr.Append(
				allErrors,
				errors.New("max retries reached"),
			)
			break
		}

		// 타이머를 생성하여 요청 타임아웃 관리
		timer := time.NewTimer(rt.requestTimeout)
		done := make(chan struct{})
		var (
			response   *http.Response
			respErr    error
			statusCode int = -1 // 응답 실패시 -1
		)

		go func() {
			// RoundTrip 호출
			response, respErr = rt.RoundTripper.RoundTrip(req)
			close(done)
		}()

		select {
		case <-timer.C:
			timeoutErr := fmt.Errorf("request timeout attempt(%d)", attempt)
			rt.debugLog(attempt, statusCode, timeoutErr)
			allErrors = multierr.Append(allErrors, timeoutErr)
		case <-done:
			timer.Stop() // 타이머 멈춤
			if response != nil {
				statusCode = response.StatusCode
			}
			shouldRetry, retryErr := rt.shouldRetry(statusCode, respErr)
			if shouldRetry {
				allErrors = multierr.Append(
					allErrors,
					errors.Wrapf(retryErr, "attempt(%d)", attempt),
				)
				rt.debugLog(attempt, statusCode, retryErr)
				time.Sleep(rt.backoffPolicy(attempt))
				continue
			}
			return response, nil
		}
	}
	return nil, allErrors

}

// shouldRetry 재시도 여부를 판단
func (rt *retriableTransport) shouldRetry(statusCode int, err error) (bool, error) {
	if err != nil {
		return true, err
	}

	if reason, shouldRetry := rt.retryStatusCodes[statusCode]; shouldRetry {
		return shouldRetry, errors.New(reason)
	}

	return false, nil
}

// debugLog 디버그 메시지를 출력
func (rt *retriableTransport) debugLog(attempt int, statusCode int, err error) {
	if rt.debugMode {
		log.Printf(
			"retrying request. Attempt: %d, StatusCode: %d, Error: %v\n",
			attempt,
			statusCode,
			err.Error(),
		)
	}
}

// extendDefault는 기본 재시도 상태 코드 맵을 확장
func extendDefault(additional []int) map[int]string {
	retryMap := make(map[int]string)
	for code, msg := range defaultRetryStatusMap {
		retryMap[code] = msg
	}
	for _, code := range additional {
		if _, exists := retryMap[code]; !exists {
			retryMap[code] = http.StatusText(code)
		}
	}
	return retryMap
}
