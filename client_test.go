package galf

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/globocom/goreq"
	"gopkg.in/check.v1"
)

type clientSuite struct {
	server *httptest.Server
}

var _ = check.Suite(&clientSuite{})

func (cs *clientSuite) SetUpSuite(c *check.C) {
	cs.server = newTestServerToken()
}

func (cs *clientSuite) TearDownSuite(c *check.C) {
	cs.server.Close()
}

func (cs *clientSuite) TestGetClient(c *check.C) {
	ts := newTestServerCustom(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"method": "GET"}`)
	})
	defer ts.Close()

	client := NewClient()
	url := fmt.Sprintf("%s/get/feed/1", ts.URL)
	resp, err := client.Get(url)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)

	body, _ := resp.Body.ToString()
	c.Assert(body, check.Equals, `{"method": "GET"}`)
}

// nolint:dupl
func (cs *clientSuite) TestPostClient(c *check.C) {
	ts := newTestServerCustom(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"method": "POST"}`)
	})
	defer ts.Close()

	assertPost := func(resp *goreq.Response, err error) {
		c.Assert(err, check.IsNil)
		c.Assert(resp.StatusCode, check.Equals, http.StatusCreated)

		body, _ := resp.Body.ToString()
		c.Assert(body, check.Equals, `{"method": "POST"}`)
	}

	client := NewClient()
	url := fmt.Sprintf("%s/post/feed/1", ts.URL)

	// body post == nil
	resp, err := client.Post(url, nil)
	assertPost(resp, err)

	// body post == io.Reader
	bodyReader := strings.NewReader(`{"body": "test"}`)
	resp, err = client.Post(url, bodyReader)
	assertPost(resp, err)

	// body post == string
	bodyString := "{'bodyPost': 'test'}"
	resp, err = client.Post(url, bodyString)
	assertPost(resp, err)

	// body post == []byte
	bodyBytes := []byte("{'bodyPost': 'test'}")
	resp, err = client.Post(url, bodyBytes)
	assertPost(resp, err)

	// body post == struct
	type bodySt struct {
		BodyPost string `json:"bodyPost"`
	}
	bodyStruct := bodySt{"test"}
	resp, err = client.Post(url, bodyStruct)
	assertPost(resp, err)
}

// nolint:dupl
func (cs *clientSuite) TestPutClient(c *check.C) {
	ts := newTestServerCustom(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"method": "PUT"}`)
	})
	defer ts.Close()

	assertPut := func(resp *goreq.Response, err error) {
		c.Assert(err, check.IsNil)
		c.Assert(resp.StatusCode, check.Equals, http.StatusOK)

		body, _ := resp.Body.ToString()
		c.Assert(body, check.Equals, `{"method": "PUT"}`)
	}

	client := NewClient()
	url := fmt.Sprintf("%s/put/feed/1", ts.URL)

	// body put == nil
	resp, err := client.Put(url, nil)
	assertPut(resp, err)

	// body put == io.Reader
	bodyReader := strings.NewReader(`{"bodyPut": "test"}`)
	resp, err = client.Put(url, bodyReader)
	assertPut(resp, err)

	// body put == string
	bodyString := "{'bodyPut': 'test'}"
	resp, err = client.Put(url, bodyString)
	assertPut(resp, err)

	// body put == []byte
	bodyBytes := []byte("{'bodyPut': 'test'}")
	resp, err = client.Put(url, bodyBytes)
	assertPut(resp, err)

	// body put == struct
	type bodySt struct {
		BodyPut string `json:"bodyPut"`
	}
	bodyStruct := bodySt{"test"}
	resp, err = client.Put(url, bodyStruct)
	assertPut(resp, err)
}

func (cs *clientSuite) TestDeleteClient(c *check.C) {
	ts := newTestServerCustom(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	client := NewClient()
	url := fmt.Sprintf("%s/delete/feed/1", ts.URL)
	resp, err := client.Delete(url)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusNoContent)

	body, _ := resp.Body.ToString()
	c.Assert(body, check.Equals, "")
}

func (cs *clientSuite) TestStatusUnauthorizedClient(c *check.C) {
	ts := newTestServerCustom(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	defer ts.Close()

	client := NewClient()
	url := fmt.Sprintf("%s/StatusUnauthorized/feed/1", ts.URL)
	resp, err := client.Get(url)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusUnauthorized)
}

func (cs *clientSuite) TestDefaultClientOptionsClient(c *check.C) {
	ts := newTestServerCustom(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", r.Header.Get("Content-Type")+"; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	client := NewClient()
	url := fmt.Sprintf("%s/ClientOptions/feed/1", ts.URL)
	resp, err := client.Get(url)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, "application/json; charset=utf-8")

	clientOptions := ClientOptions{
		Timeout:       DefaultClientTimeout,
		MaxRetries:    DefaultClientMaxRetries,
		Backoff:       ConstantBackOff,
		ShowDebug:     false,
		HystrixConfig: nil,
	}

	client = NewClient(clientOptions)
	url = fmt.Sprintf("%s/ClientOptions/feed/1", ts.URL)
	resp, err = client.Get(url)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, "application/json; charset=utf-8")
}

func (cs *clientSuite) TestClientOptionsClient(c *check.C) {
	ts := newTestServerCustom(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", r.Header.Get("Content-Type")+"; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	clientOptions := ClientOptions{
		ContentType:   "application/my-custom-type",
		Timeout:       DefaultClientTimeout,
		MaxRetries:    DefaultClientMaxRetries,
		Backoff:       ConstantBackOff,
		ShowDebug:     false,
		HystrixConfig: nil,
	}

	client := NewClient(clientOptions)
	url := fmt.Sprintf("%s/ClientOptions/feed/1", ts.URL)
	resp, err := client.Get(url)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, "application/my-custom-type; charset=utf-8")
}

func (cs *clientSuite) TestHystrixClient(c *check.C) {
	ts := newTestServerCustom(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"hystrix": "OK"}`)
	})
	defer ts.Close()

	hystrixConfig := hystrix.CommandConfig{
		Timeout:                5000,
		SleepWindow:            2000,
		RequestVolumeThreshold: 50,
		MaxConcurrentRequests:  100,
	}

	hystrixConfigName := "hystrixConfigName"
	HystrixConfigureCommand(hystrixConfigName, hystrixConfig)
	clientOptions := NewClientOptions(
		DefaultClientTimeout,
		false,
		DefaultClientMaxRetries,
		hystrixConfigName,
	)

	client := NewClient(clientOptions)
	url := fmt.Sprintf("%s/hystrixconfig/feed/1", ts.URL)
	resp, err := client.Get(url)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)

	body, _ := resp.Body.ToString()
	c.Assert(body, check.Equals, `{"hystrix": "OK"}`)
}

func (cs *clientSuite) TestHystrixConfigTimeoutClient(c *check.C) {
	timeout := 200

	ts := newTestServerCustom(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(timeout+10) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"hystrix": "OK"}`)
	})
	defer ts.Close()

	hystrixConfig := hystrix.CommandConfig{
		Timeout:                timeout,
		SleepWindow:            2000,
		RequestVolumeThreshold: 50,
		MaxConcurrentRequests:  100,
	}

	hystrixConfigName := "hystrixConfigTimeout"
	HystrixConfigureCommand(hystrixConfigName, hystrixConfig)
	clientOptions := NewClientOptions(
		DefaultClientTimeout,
		false,
		DefaultClientMaxRetries,
		hystrixConfigName,
	)

	client := NewClient(clientOptions)
	url := fmt.Sprintf("%s/hystrixconfigtimeout/feed/1", ts.URL)
	resp, err := client.Get(url)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "hystrix: timeout")
	c.Assert(resp, check.IsNil)
}

func (cs *clientSuite) TestHystrixConfigNotFoundClient(c *check.C) {
	hystrixConfigName := "hystrixConfigTimeout"
	clientOptions := NewClientOptions(
		DefaultClientTimeout,
		false,
		DefaultClientMaxRetries,
		hystrixConfigName,
	)

	client := NewClient(clientOptions)
	resp, err := client.Get("/hystrixconfignotfound/feed/1")
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "Hystrix config name not found: "+hystrixConfigName)
	c.Assert(resp, check.IsNil)
}

func (cs *clientSuite) TestGetClientRequestOptionsHeader(c *check.C) {

	ts := newTestServerCustom(func(rw http.ResponseWriter, r *http.Request) {
		responseHeaders(rw, r)
		rw.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	url := fmt.Sprintf("%s/requestOptions/feed/1", ts.URL)

	requestOptions := NewRequestOptions()
	requestOptions.AddHeader("token", "1234567890")

	client := NewClient()
	resp, err := client.Get(url, requestOptions)

	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Header.Get("token"), check.Equals, "1234567890")
}

func (cs *clientSuite) TestGetClientRequestOptionsHeaders(c *check.C) {

	ts := newTestServerCustom(func(rw http.ResponseWriter, r *http.Request) {
		responseHeaders(rw, r)
		rw.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	url := fmt.Sprintf("%s/requestOptions/feed/2", ts.URL)

	headers := map[string]string{
		"header1": "123",
		"header2": "456",
	}
	requestOptions := NewRequestOptions()
	requestOptions.AddHeaders(headers)

	client := NewClient()
	resp, err := client.Get(url, requestOptions)

	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Header.Get("header1"), check.Equals, "123")
	c.Assert(resp.Header.Get("header2"), check.Equals, "456")
}

func (cs *clientSuite) TestConcurrencyRequest(c *check.C) {
	ts := newTestServerCustom(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"method": "GET"}`)
	})
	defer ts.Close()

	client := NewClient()
	url := fmt.Sprintf("%s/get/feed/1", ts.URL)
	token, e := client.TokenManager.GetToken()
	c.Assert(e, check.IsNil)

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resp, err := client.request(token.Authorization, http.MethodGet, url, nil, nil)

			c.Assert(err, check.IsNil)
			c.Assert(resp.StatusCode, check.Equals, http.StatusOK)

			body, _ := resp.Body.ToString()
			c.Assert(body, check.Equals, `{"method": "GET"}`)
		}()
	}
	wg.Wait()

}

func responseHeaders(rw http.ResponseWriter, r *http.Request) {
	for name, headers := range r.Header {
		for _, v := range headers {
			rw.Header().Set(name, v)
		}
	}
}
