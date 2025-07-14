# monkey patching in testing

## what is monkey patching?

monkey patching is a technique that allows you to dynamically modify or extend functions, methods, and classes at runtime. in the context of testing, it's used to replace real implementations with mock versions, enabling you to test code without external dependencies like api calls, database connections, or file system operations.

the core concept involves temporarily replacing a function or method with a substitute that returns predictable results, allowing you to test your code in isolation.

## why use monkey patching in tests?

### advantages
- **isolate dependencies**: test code without relying on external services
- **control behavior**: make functions return predictable results
- **continue development**: work on applications even when external apis aren't available
- **fast tests**: avoid slow network calls or database operations
- **deterministic testing**: ensure tests produce consistent results

### when to use monkey patching
- testing functions that call external apis
- mocking time-dependent functions
- stubbing file system operations
- replacing database calls during unit tests
- testing error scenarios by forcing functions to fail

## general monkey patching concepts

### basic principles
1. **temporarily replace** the original function with a mock
2. **maintain the same signature** as the original function
3. **restore the original** function after the test
4. **return predictable data** that matches the expected structure

### common patterns
- patch at the beginning of a test
- define mock behavior
- run the test
- restore original functionality
- verify expectations

## monkey patching in go

go presents unique challenges for monkey patching due to its compiled nature and type system. however, several approaches and libraries make it possible.

### challenges in go
- **compiled language**: functions are compiled, making runtime modification difficult
- **type safety**: strict typing requires exact signature matching
- **inlining**: go compiler may inline functions, preventing patching
- **thread safety**: concurrent access during patching can cause issues

### solutions and libraries

#### 1. gomonkey library

the most popular library for monkey patching in go is [gomonkey](https://github.com/agiledragon/gomonkey).

**installation:**
```bash
# for versions below v2.1.0
go get github.com/agiledragon/gomonkey@v2.0.2

# for v2.1.0 and above
go get github.com/agiledragon/gomonkey/v2@v2.11.0
```

**features:**
- patch functions
- patch public and private methods
- patch interfaces
- patch function variables
- patch global variables
- support for sequential patches

**basic usage:**
```go
import (
    "github.com/agiledragon/gomonkey/v2"
    "testing"
)

func TestWithMonkeyPatch(t *testing.T) {
    // patch a function
    patches := gomonkey.ApplyFunc(originalFunction, func(param string) string {
        return "mocked result"
    })
    defer patches.Reset()
    
    // run your test
    result := functionUnderTest()
    // assertions...
}
```

**important considerations:**
- must run tests with `-gcflags=all=-l` to disable inlining
- not thread-safe
- supports specific architectures (amd64, arm64, 386, loong64, riscv64)
- works on linux, macos, and windows

#### 2. bouke/monkey library

an alternative library that provides low-level monkey patching capabilities.

**installation:**
```bash
go get bou.ke/monkey
```

**usage:**
```go
import (
    "bou.ke/monkey"
    "testing"
    "time"
)

func TestWithMonkey(t *testing.T) {
    // patch time.Now to return a fixed time
    fixedTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
    monkey.Patch(time.Now, func() time.Time {
        return fixedTime
    })
    defer monkey.UnpatchAll()
    
    // run your test
    result := getCurrentTimeString()
    expected := "2023-01-01T00:00:00Z"
    if result != expected {
        t.Errorf("expected %s, got %s", expected, result)
    }
}
```

**limitations:**
- archived project (no longer maintained)
- only works on linux and macos
- may not work on security-oriented systems
- requires `-gcflags=-l` flag

### practical examples

#### patching external api calls

```go
package main

import (
    "encoding/json"
    "net/http"
    "testing"
    "github.com/agiledragon/gomonkey/v2"
)

type apiResponse struct {
    Status string `json:"status"`
    Data   string `json:"data"`
}

func makeAPICall(url string) (*apiResponse, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result apiResponse
    err = json.NewDecoder(resp.Body).Decode(&result)
    return &result, err
}

func TestAPICall(t *testing.T) {
    // patch http.Get to return a mock response
    patches := gomonkey.ApplyFunc(http.Get, func(url string) (*http.Response, error) {
        // return a mock response
        mockResponse := &apiResponse{
            Status: "ok",
            Data:   "test data",
        }
        // create a mock http.Response with the json data
        // ... implementation details
        return mockHTTPResponse, nil
    })
    defer patches.Reset()
    
    result, err := makeAPICall("https://api.example.com/data")
    if err != nil {
        t.Fatal(err)
    }
    
    if result.Status != "ok" {
        t.Errorf("expected status 'ok', got %s", result.Status)
    }
}
```

#### patching methods

```go
type dataService struct {
    client *http.Client
}

func (d *dataService) fetchData(id string) (string, error) {
    // implementation that makes http request
    return "real data", nil
}

func TestDataService(t *testing.T) {
    service := &dataService{}
    
    // patch the method
    patches := gomonkey.ApplyMethod(reflect.TypeOf(service), "fetchData", 
        func(_ *dataService, id string) (string, error) {
            return "mocked data", nil
        })
    defer patches.Reset()
    
    result, err := service.fetchData("123")
    if err != nil {
        t.Fatal(err)
    }
    
    if result != "mocked data" {
        t.Errorf("expected 'mocked data', got %s", result)
    }
}
```

### best practices for go monkey patching

1. **disable inlining**: always run tests with `-gcflags=all=-l`
2. **clean up patches**: use `defer patches.Reset()` or `defer monkey.UnpatchAll()`
3. **avoid in production**: only use monkey patching in test code
4. **prefer interfaces**: when possible, use dependency injection with interfaces
5. **document patches**: clearly comment what you're patching and why
6. **test isolation**: ensure patches don't affect other tests

### running tests with monkey patching

```bash
# run tests with inlining disabled
go test -gcflags=all=-l

# run specific test
go test -gcflags=all=-l -run TestSpecificFunction

# run with verbose output
go test -gcflags=all=-l -v
```

## alternative approaches

### dependency injection
instead of monkey patching, consider using dependency injection with interfaces:

```go
type httpClient interface {
    Get(url string) (*http.Response, error)
}

type service struct {
    client httpClient
}

func (s *service) getData(url string) (string, error) {
    resp, err := s.client.Get(url)
    // ... implementation
}

// in tests, inject a mock client
func TestService(t *testing.T) {
    mockClient := &mockHTTPClient{}
    service := &service{client: mockClient}
    
    // test with mock client
    result, err := service.getData("test-url")
    // assertions...
}
```

### function variables
for simple functions, you can use function variables:

```go
var timeNow = time.Now

func getCurrentTime() time.Time {
    return timeNow()
}

func TestGetCurrentTime(t *testing.T) {
    originalTimeNow := timeNow
    defer func() { timeNow = originalTimeNow }()
    
    fixedTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
    timeNow = func() time.Time {
        return fixedTime
    }
    
    result := getCurrentTime()
    if !result.Equal(fixedTime) {
        t.Errorf("expected %v, got %v", fixedTime, result)
    }
}
```

## comparison with other languages

### python
python makes monkey patching straightforward with its dynamic nature:

```python
import unittest
from unittest.mock import patch

class TestExample(unittest.TestCase):
    @patch('requests.get')
    def test_api_call(self, mock_get):
        mock_get.return_value.json.return_value = {'status': 'ok'}
        result = make_api_call()
        self.assertEqual(result['status'], 'ok')
```

### javascript
javascript also supports easy monkey patching:

```javascript
const originalFetch = global.fetch;
global.fetch = jest.fn(() => 
    Promise.resolve({
        json: () => Promise.resolve({status: 'ok'})
    })
);

// run tests
// restore
global.fetch = originalFetch;
```

## testing strategies

### unit test patterns
1. **arrange**: set up the monkey patch
2. **act**: execute the function under test
3. **assert**: verify the expected behavior
4. **cleanup**: restore original functionality

### integration considerations
- use monkey patching sparingly in integration tests
- prefer real implementations when testing system interactions
- use mocks for external dependencies only

### test organization
```go
func TestSuite(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        patch    func()
        cleanup  func()
    }{
        {
            name:     "successful api call",
            input:    "test-id",
            expected: "success",
            patch: func() {
                // apply patches
            },
            cleanup: func() {
                // cleanup patches
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.patch != nil {
                tt.patch()
            }
            if tt.cleanup != nil {
                defer tt.cleanup()
            }
            
            result := functionUnderTest(tt.input)
            if result != tt.expected {
                t.Errorf("expected %s, got %s", tt.expected, result)
            }
        })
    }
}
```

## common pitfalls and solutions

### pitfall 1: inlining issues
**problem**: patches don't work because functions are inlined
**solution**: use `-gcflags=all=-l` flag

### pitfall 2: race conditions
**problem**: concurrent tests interfere with each other
**solution**: avoid global patches, use test-specific patches

### pitfall 3: forgotten cleanup
**problem**: patches affect other tests
**solution**: always use `defer` for cleanup

### pitfall 4: signature mismatches
**problem**: mock function signature doesn't match original
**solution**: carefully match parameter types and return values

## conclusion

monkey patching is a powerful technique for testing, especially when dealing with external dependencies. while go presents some challenges due to its compiled nature, libraries like gomonkey make it feasible. however, it should be used judiciously, with preference given to dependency injection and interface-based approaches when possible.

remember to:
- use monkey patching primarily for testing external dependencies
- always clean up patches after tests
- prefer interfaces and dependency injection when designing code
- disable inlining when running tests with monkey patches
- document your patches clearly

## references

1. wesselhuising.medium.com - "use monkey patching to continue developing applications depending on external api calls while waiting for access" - https://wesselhuising.medium.com/use-monkey-patching-to-continue-developing-applications-depending-on-external-api-calls-while-931c11ade11b

2. github.com/agiledragon/gomonkey - "gomonkey is a library to make monkey patching in unit tests easy" - https://github.com/agiledragon/gomonkey

3. medium.com/@andrewdavisescalona - "testing in go — some tools you can use" - https://medium.com/@andrewdavisescalona/testing-in-go-some-tools-you-can-use-f3e79b398d8d
