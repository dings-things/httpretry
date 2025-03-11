package httpretry

import "time"

type (
	// HTTPOption http 설정
	HTTPOption func(*Settings)
)

// NewHTTPSettings httpOption들을 인자로 설정 struct을 생성
//
// fx 또는 env를 사용하지 않고, HTTP Retry Client를 사용하고자 할 때, 옵션 세팅을 위한 생성자
//
// Parameters:
//   - opts: (...HTTPOption) 기본 세팅 값에서 변경이 필요한 경우, 추가되는 옵션 값
func NewHTTPSettings(opts ...HTTPOption) *Settings {
	// 기본 설정 적용
	settings := &Settings{
		MaxRetry:              3,
		DebugMode:             false,
		Insecure:              true,
		MaxIdleConns:          15,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		RequestTimeout:        10 * time.Second,
		BackoffPolicy:         defaultBackoffPolicy,
	}

	// Option 함수들을 실행하여 설정 적용
	for _, opt := range opts {
		opt(settings)
	}
	return settings
}

// WithMaxRetry MaxRetry 설정을 변경하는 Option
//
// 최대로 재시도하는 횟수를 지정합니다. 실패 시 지정된 횟수 +1회 만큼 재시도합니다.
//
// Parameters:
//   - maxRetry: (int) 재시도 하고자 하는 횟수
func WithMaxRetry(maxRetry int) HTTPOption {
	return func(s *Settings) {
		s.MaxRetry = maxRetry
	}
}

// WithDebugMode 디버그 모드 설정을 변경하는 Option
//
// 디버그 모드 여부를 확인 합니다. 디버그 모드 전환 시, stdout으로 재시도 사유를 프린트합니다.
//
// Parameters:
//   - isDebug: (bool) 디버깅 모드 여부
func WithDebugMode(isDebug bool) HTTPOption {
	return func(s *Settings) {
		s.DebugMode = isDebug
	}
}

// WithInsecure SSL/TLS 인증서 유효성 검증 여부 설정을 변경하는 Option
//
// 인증서 유효성 검사를 실시 할 지 여부를 확인합니다. true 시, 인증서 유효성 검사를 거치지 않습니다.
//
// Parameters:
//   - isSecure: (bool) SSL/TLS 인증서 유효성 검증 여부
func WithInsecure(insecure bool) HTTPOption {
	return func(s *Settings) {
		s.Insecure = insecure
	}
}

// WithIdleConnTimeout IdleConnTimeout 설정을 변경하는 Option
//
// 클라이언트가 유휴 상태인 TCP 연결을 얼마나 유지할 것인지 결정합니다.
//
// Parameters:
//   - timeout: (time.Duration) TCP 연결 유지 시간
func WithIdleConnTimeout(timeout time.Duration) HTTPOption {
	return func(s *Settings) {
		s.IdleConnTimeout = timeout
	}
}

// WithTLSHandshakeTimeout TLSHandshakeTimeout 설정을 변경하는 Option
//
// 최초 SSL/TLS 인증서 검증이후, TLS 핸드셰이크가 유지되는 시간을 결정합니다.
//
// Parameters:
//   - timeout: (time.Duration) TLS 핸드셰이크 유지 시간
func WithTLSHandshakeTimeout(timeout time.Duration) HTTPOption {
	return func(s *Settings) {
		s.TLSHandshakeTimeout = timeout
	}
}

// WithExpectContinueTimeout ExpectContinueTimeout 설정을 변경하는 Option
//
// 클라이언트가 서버에 데이터를 전송하기 전에 서버의 승인을 기다리는 시간
//
// 주의: 1초 이하의 단위로 timeout을 거는 것은 위험합니다. 서버가 승인을 보내지 않을 경우, 클라이언트는 요청을 취소합니다.
//
// Parameters:
//   - timeout: (time.Duration) 승인 대기 시간
func WithExpectContinueTimeout(timeout time.Duration) HTTPOption {
	return func(s *Settings) {
		s.ExpectContinueTimeout = timeout
	}
}

// WithResponseHeaderTimeout ResponseHeaderTimeout 설정을 변경하는 Option
//
// 클라이언트가 서버로부터 응답 헤더를 받는 데까지 걸리는 최대 시간을 지정
//
// 주의: 1초 이하의 단위로 timeout을 거는 것은 위험합니다. 서버가 응답 헤더를 보내지 않을 경우, 클라이언트는 요청을 취소합니다.
//
// Parameters:
//   - timeout: (time.Duration) 응답 헤더 수신 시간
func WithResponseHeaderTimeout(timeout time.Duration) HTTPOption {
	return func(s *Settings) {
		s.ResponseHeaderTimeout = timeout
	}
}

// WithRequestTimeout RequestTimeout 설정을 변경하는 Option
//
// 전체 HTTP요청의 최대 실행 시간을 지정. Timeout 발생 시, 전체 요청을 취소하기 때문에 재시도 하지 않음 주의
//
// 주의: 1초 이하의 단위로 timeout을 거는 것은 위험합니다. 서버가 응답 헤더를 보내지 않을 경우, 클라이언트는 요청을 취소합니다.
//
// # REFS
//   - https://uptrace.dev/blog/golang-context-timeout.html
//   - https://devblogs.microsoft.com/premier-developer/the-art-of-http-connection-pooling-how-to-optimize-your-connections-for-peak-performance/#create-your-own-keep-alive-strategy-or-not
//
// Parameters:
//   - timeout: (time.Duration) 전체 요청 최대 실행 시간
func WithRequestTimeout(timeout time.Duration) HTTPOption {
	return func(s *Settings) {
		s.RequestTimeout = timeout
	}
}

// WithMaxIdleConns MaxIdleConns 설정을 변경하는 Option
//
// 클라이언트가 유지할 수 있는 최대 유휴(Idle) 연결의 수를 지정
//
// Parameters:
//   - maxIdleConns: (int) TCP 연결의 최대 수
func WithMaxIdleConns(maxIdleConns int) HTTPOption {
	return func(s *Settings) {
		s.MaxIdleConns = maxIdleConns
	}
}

// WithBackoffPolicy 요청 실패 시, backoff 정책을 변경하는 Option
//
// 기본으로 지수 백오프가 적용됨
//
// Parameters:
//   - policy: (func(attempt int) time.Duration) 백오프 정책
func WithBackoffPolicy(policy func(attempt int) time.Duration) HTTPOption {
	return func(s *Settings) {
		s.BackoffPolicy = policy
	}
}

// 기본 백오프 정책 (지수 백오프)
func defaultBackoffPolicy(attempt int) time.Duration {
	return time.Duration(1<<attempt) * time.Second
}
