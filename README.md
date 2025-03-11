# httpretry

`httpretry` is a robust and configurable HTTP client for Go that enhances the standard `http.Client` by providing **automatic retries for requests failing due to timeouts** or other transient errors. It allows developers to **retain their existing HTTP client logic** while adding powerful retry mechanisms with customizable backoff policies.

## Features

- 🛠 **Enhances existing HTTP clients** with retry capabilities
- 🔄 **Automatic retries for timeout-related errors** and transient HTTP failures
- ⏳ **Customizable timeouts and connection settings**
- 📜 **Configurable backoff strategies** (exponential, linear, or custom)
- 🔍 **Debug mode** for logging retry attempts
- 🛠 **Seamless integration** with both environment-based and programmatic configurations

---

## Installation

To install `httpretry`, run:

```sh
go get github.com/dings-things/httpretry
```

---

## Quickstart

Here’s how to **use httpretry without modifying existing HTTP logic**:

```go
package main

import (
	"fmt"
	"net/http"
	"github.com/dings-things/httpretry"
)

func main() {
	// Create retry-enabled settings
	settings := httpretry.NewHTTPSettings(
		httpretry.WithMaxRetry(3),
		httpretry.WithDebugMode(true),
		httpretry.WithRequestTimeout(5 * time.Second),
	)

	// Create an HTTP client with retry support
	client := httpretry.NewClient(settings)

	// Making a request that may fail due to timeouts or transient errors
	resp, err := client.Get("https://httpbin.org/status/500")

	if err != nil {
		fmt.Println("Request failed after retries:", err)
	} else {
		fmt.Println("Response received with status:", resp.StatusCode)
	}
}
```

---

## Why `httpretry`?

**Traditional HTTP clients immediately fail on timeouts or transient errors.**  
With `httpretry`, you can **extend your existing client logic** to automatically retry failed requests caused by:

✅ Network timeouts  
✅ Gateway or service unavailable errors (5xx responses)  
✅ Temporary connection failures  

It ensures that your application remains **resilient and responsive** even in unreliable network conditions.

---

## Usage

### Creating a Custom HTTP Client with Retry Support

```go
settings := httpretry.NewHTTPSettings(
    httpretry.WithMaxRetry(5),
    httpretry.WithRequestTimeout(10 * time.Second), // Handles timeout-based retries
    httpretry.WithBackoffPolicy(func(attempt int) time.Duration {
        return time.Duration(attempt) * time.Second
    }),
)

client := httpretry.NewClient(settings)
```

### Making HTTP Requests with Automatic Retries

Instead of modifying your HTTP logic, just use the `client` as you normally would:

```go
req, _ := http.NewRequest("GET", "https://httpbin.org/status/502", nil)
resp, err := client.Do(req)

if err != nil {
    fmt.Println("Request failed after retries:", err)
} else {
    fmt.Println("Response received:", resp.StatusCode)
}
```

### Handling Requests That May Timeout

If you have long-running requests, you can use **timeouts with retries**:

```go
settings := httpretry.NewHTTPSettings(
    httpretry.WithMaxRetry(3),
    httpretry.WithRequestTimeout(5 * time.Second), // Retries only when timeout occurs
)

client := httpretry.NewClient(settings)
```

Now, requests that time out will **automatically retry** instead of failing instantly.

---

## Configuring via Environment Variables

You can configure `httpretry` using environment variables:

```sh
export MAX_REQUEST_RETRY=5
export DEBUG_MODE=true
export REQUEST_TIMEOUT=5s
```

Then initialize settings using:

```go
settings := httpretry.NewSettings()
client := httpretry.NewClient(settings)
```

---

## Advanced Configuration

#### Set Maximum Retries
```go
settings := httpretry.NewHTTPSettings(
    httpretry.WithMaxRetry(5),
)
```

#### Enable Debug Mode
```go
settings := httpretry.NewHTTPSettings(
    httpretry.WithDebugMode(true),
)
```

#### Customize Backoff Policy
```go
settings := httpretry.NewHTTPSettings(
    httpretry.WithBackoffPolicy(func(attempt int) time.Duration {
        return time.Duration(attempt*2) * time.Second
    }),
)
```

#### Control Timeout Handling
```go
settings := httpretry.NewHTTPSettings(
    httpretry.WithRequestTimeout(5 * time.Second),
)
```

---

## License

`httpretry` is open-source and available under the MIT License.