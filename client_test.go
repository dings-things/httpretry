package httpretry_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dings-things/httpretry"
	"github.com/stretchr/testify/assert"
)

func TestHTTPRetryClientRetry(t *testing.T) {
	t.Run("요청 성공 시, 응답 반환 테스트", func(t *testing.T) {
		// given
		testServer := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Success response"))
			}),
		)
		defer testServer.Close()

		retryClient := httpretry.NewClient(
			httpretry.NewHTTPSettings(
				httpretry.WithDebugMode(true),
				httpretry.WithRequestTimeout(1*time.Second),
				httpretry.WithMaxRetry(3),
			),
		)

		// when
		resp, err := retryClient.Get(testServer.URL)

		// then
		assert.Nil(t, err, "에러가 발생하지 않아야 합니다.")
		assert.NotNil(t, resp, "응답이 있어야 합니다.")
	})

	t.Run("request timeout 초과 시, 재시도 테스트", func(t *testing.T) {
		// given
		testServer := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second) // 응답 지연
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Delayed response"))
			}),
		)
		defer testServer.Close()

		retryClient := httpretry.NewClient(
			httpretry.NewHTTPSettings(
				httpretry.WithDebugMode(true),
				httpretry.WithRequestTimeout(1*time.Second),
				httpretry.WithMaxRetry(3),
			),
		)

		// when
		resp, err := retryClient.Get(testServer.URL)

		// then
		assert.NotNil(t, err, "timeout 에러가 발생해야 합니다.")
		assert.Nil(t, resp, "응답이 없어야 합니다.")
		assert.ErrorContains(
			t,
			err,
			"max retries reached",
			"재시도 횟수 초과 에러 메시지가 출력되어야 합니다.",
		)
	})

	t.Run("retry status code 재시도 이후 성공 테스트", func(t *testing.T) {
		// given
		reqCount := 0
		testServer := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reqCount++

				if reqCount > 2 {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Success response"))
				} else {
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte("Service unavailable"))
				}
			}),
		)
		defer testServer.Close()

		retryClient := httpretry.NewClient(
			httpretry.NewHTTPSettings(
				httpretry.WithDebugMode(true),
				httpretry.WithRequestTimeout(1*time.Second),
				httpretry.WithMaxRetry(3),
			),
			http.StatusServiceUnavailable,
		)

		// when
		resp, err := retryClient.Get(testServer.URL)

		// then
		assert.Nil(t, err, "에러가 발생하지 않아야 합니다.")
		assert.NotNil(t, resp, "응답이 있어야 합니다.")
	})

	t.Run("retry status code 등록 시, 재시도 테스트", func(t *testing.T) {
		// given
		testServer := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Service unavailable"))
			}),
		)
		defer testServer.Close()

		retryClient := httpretry.NewClient(
			httpretry.NewHTTPSettings(
				httpretry.WithDebugMode(true),
				httpretry.WithRequestTimeout(1*time.Second),
				httpretry.WithMaxRetry(3),
			),
			http.StatusNotFound,
		)

		// when
		resp, err := retryClient.Get(testServer.URL)

		// then
		assert.NotNil(t, err, "에러가 발생해야 합니다.")
		assert.Nil(t, resp, "응답이 있어야 합니다.")
		assert.ErrorContains(
			t,
			err,
			http.StatusText(http.StatusNotFound),
			"재시도 상태 코드 메시지가 출력되어야 합니다.",
		)
	})
}

func TestRetriableTransport_ParentContextCancel(t *testing.T) {
	t.Run("부모 context가 timeout으로 deadline exceeeded인 경우, 재시도 하지 않고 에러 반환 테스트", func(t *testing.T) {
		// given
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(300 * time.Millisecond) // 300ms 지연 후 응답
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		retryClient := httpretry.NewClient(
			httpretry.NewHTTPSettings(
				httpretry.WithDebugMode(true),
				httpretry.WithRequestTimeout(1*time.Second),
				httpretry.WithMaxRetry(3),
			),
			http.StatusNotFound,
		)

		// 100ms 후에 만료되는 context 생성
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		assert.NoError(t, err)

		// when
		_, err = retryClient.Do(req)

		// then
		assert.Error(t, err)
		assert.ErrorContains(t, err, "cancelled from parent context")
	})

	t.Run("부모 context가 cancel로 인해 취소된 경우, 재시도 하지 않고 에러 반환 테스트", func(t *testing.T) {
		// given
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(300 * time.Millisecond) // 300ms 지연 후 응답
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		retryClient := httpretry.NewClient(
			httpretry.NewHTTPSettings(
				httpretry.WithDebugMode(true),
				httpretry.WithRequestTimeout(1*time.Second),
				httpretry.WithMaxRetry(3),
			),
			http.StatusNotFound,
		)

		// 취소되는 context 생성
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		assert.NoError(t, err)

		// when
		_, err = retryClient.Do(req)

		// then
		assert.Error(t, err)
		assert.ErrorContains(t, err, "cancelled from parent context")
	})

	t.Run("context canceled during body read", func(t *testing.T) {
		// given
		testServer := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush() // 응답의 헤더를 전송하여 클라이언트가 body를 읽도록 유도
				w.Write([]byte("This part won't be read"))
			}),
		)
		defer testServer.Close()

		retryClient := httpretry.NewClient(
			httpretry.NewHTTPSettings(
				httpretry.WithDebugMode(true),
				httpretry.WithRequestTimeout(500*time.Millisecond), // 타임아웃 짧게 설정
				httpretry.WithMaxRetry(1),
			),
		)

		req, _ := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			testServer.URL,
			nil,
		)

		// when
		resp, err := retryClient.Do(req)

		// then
		assert.Nil(t, err, "응답 자체는 성공해야 합니다.")
		assert.NotNil(t, resp, "응답이 있어야 합니다.")

		respBody, readErr := io.ReadAll(resp.Body)
		t.Log(string(respBody))
		assert.NoError(t, readErr, "응답 body를 읽는데 에러가 발생하지 않아야 합니다.")
	})
}
