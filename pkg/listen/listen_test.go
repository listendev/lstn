// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2023 The listen.dev team <engineering@garnet.ai>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package listen

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	internaltesting "github.com/listendev/lstn/internal/testing"
	"github.com/listendev/lstn/pkg/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/npm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	localEndpoint    = "http://127.0.0.1:3000"
	nonLocalEndpoint = "https://smtg.listen.dev"
)

type mockContextLocalEndpoint struct{}

func (ctx mockContextLocalEndpoint) Deadline() (deadline time.Time, ok bool) {
	return deadline, ok
}

func (ctx mockContextLocalEndpoint) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)

	return ch
}

func (ctx mockContextLocalEndpoint) Err() error {
	return nil
}

func (ctx mockContextLocalEndpoint) Value(key interface{}) interface{} {
	// pkgcontext.ConfigKey
	if key == pkgcontext.ConfigKey {
		c, _ := flags.NewConfigFlags()
		c.Endpoint = localEndpoint

		return c
	}

	return nil
}

type fakeRequest struct{}

func (fr *fakeRequest) IsRequest() bool {
	return true
}

func TestUnsupportedRequestType(t *testing.T) {
	o, e := newOptions(WithContext(context.Background()))
	assert.Nil(t, e)

	_, err := getEndpointURLFromContext(&fakeRequest{}, o)
	if assert.Error(t, err) {
		assert.Equal(t, "unsupported request type", err.Error())
	}
}

func TestMissingConfigurationOptions(t *testing.T) {
	o, e := newOptions(WithContext(context.Background()))
	assert.Nil(t, e)

	if _, err := getEndpointURLFromContext(&AnalysisRequest{}, o); assert.Error(t, err) {
		assert.Equal(t, "couldn't obtain configuration options", err.Error())
	}

	if _, err := getEndpointURLFromContext(&VerdictsRequest{}, o); assert.Error(t, err) {
		assert.Equal(t, "couldn't obtain configuration options", err.Error())
	}
}

func TestLocalEndpoint(t *testing.T) {
	o, e := newOptions(WithContext(mockContextLocalEndpoint{}))
	assert.Nil(t, e)

	endpointAnalysis, err := getEndpointURLFromContext(&AnalysisRequest{}, o)
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(endpointAnalysis, localEndpoint))
	assert.Equal(t, "/api/npm", getAPIPrefix(endpointAnalysis))
	assert.True(t, strings.HasSuffix(endpointAnalysis, "analysis"))

	endpointVerdicts, err := getEndpointURLFromContext(&VerdictsRequest{}, o)
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(endpointVerdicts, localEndpoint))
	assert.Equal(t, "/api/npm", getAPIPrefix(endpointVerdicts))
	assert.True(t, strings.HasSuffix(endpointVerdicts, "verdicts"))
}

type mockContextNonLocalEndpoint struct{}

func (ctx mockContextNonLocalEndpoint) Deadline() (deadline time.Time, ok bool) {
	return deadline, ok
}

func (ctx mockContextNonLocalEndpoint) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)

	return ch
}

func (ctx mockContextNonLocalEndpoint) Err() error {
	return context.DeadlineExceeded
}

func (ctx mockContextNonLocalEndpoint) Value(key interface{}) interface{} {
	// pkgcontext.ConfigKey
	if key == pkgcontext.ConfigKey {
		c, _ := flags.NewConfigFlags()
		c.Endpoint = nonLocalEndpoint

		return c
	}

	return nil
}

func TestNonLocalEndpoint(t *testing.T) {
	o, e := newOptions(WithContext(mockContextNonLocalEndpoint{}))
	assert.Nil(t, e)

	endpointAnalysis, err := getEndpointURLFromContext(&AnalysisRequest{}, o)
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(endpointAnalysis, nonLocalEndpoint))
	assert.Equal(t, "/api", getAPIPrefix(endpointAnalysis))
	assert.True(t, strings.HasSuffix(endpointAnalysis, "analysis"))

	endpointVerdicts, err := getEndpointURLFromContext(&VerdictsRequest{}, o)
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(endpointVerdicts, nonLocalEndpoint))
	assert.Equal(t, "/api", getAPIPrefix(endpointVerdicts))
	assert.True(t, strings.HasSuffix(endpointVerdicts, "verdicts"))
}

type RequestsSuite struct {
	suite.Suite
	server *httptest.Server
}

func (suite *RequestsSuite) BeforeTest(suiteName, testName string) {
	suite.Assert().Equal("RequestsSuite", suiteName)

	switch testName {
	case "TestAnalysisRequest":
		resp := []byte(`[{"name":"js-tokens","shasum":"19203fb59991df98e3a287050d4647cdeaf32499","verdicts":[],"version":"4.0.0"},{"name":"loose-envify","shasum":"71ee51fa7be4caec1a63839f7e682d8132d30caf","verdicts":[],"version":"1.4.0"}]`)
		suite.server = internaltesting.MockHTTPServer(suite.Assert(), "analysis", resp, http.StatusOK, "POST")
	case "TestVerdictsRequest":
		resp := []byte(`[{"name":"js-tokens","shasum":"19203fb59991df98e3a287050d4647cdeaf32499","verdicts":[],"version":"4.0.0"}]`)
		suite.server = internaltesting.MockHTTPServer(suite.Assert(), "verdicts", resp, http.StatusOK, "POST")
	}
}

func (suite *RequestsSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

func (suite *RequestsSuite) TestAnalysisRequestWithBackgroundContext() {
	ctx := context.Background()
	resp, data, erro := Packages(&AnalysisRequest{}, WithContext(ctx))
	suite.Assert().Nil(resp)
	suite.Assert().Nil(data)
	if suite.Assert().Error(erro) {
		suite.Assert().Equal("couldn't obtain configuration options", erro.Error())
	}
}

func (suite *RequestsSuite) TestVerdictsRequestWithBackgroundContext() {
	ctx := context.Background()
	resp, data, erro := Packages(&VerdictsRequest{}, WithContext(ctx))
	suite.Assert().Nil(resp)
	suite.Assert().Nil(data)
	if suite.Assert().Error(erro) {
		suite.Assert().Equal("couldn't obtain configuration options", erro.Error())
	}
}

func (suite *RequestsSuite) TestAnalysisRequestEmpty() {
	resp, data, erro := Packages(&AnalysisRequest{}, WithContext(mockContextLocalEndpoint{}))
	suite.Assert().Nil(resp)
	suite.Assert().Nil(data)
	if suite.Assert().Error(erro) {
		suite.Assert().Contains(erro.Error(), "package lock is mandatory")
	}
}

func (suite *RequestsSuite) TestVerdictsRequestEmpty() {
	resp, data, erro := Packages(&VerdictsRequest{}, WithContext(mockContextLocalEndpoint{}))
	suite.Assert().Nil(resp)
	suite.Assert().Nil(data)
	if suite.Assert().Error(erro) {
		suite.Assert().Contains(erro.Error(), "name is mandatory")
	}
}

func (suite *RequestsSuite) TestAnalysisRequest() {
	plj, _ := npm.NewPackageLockJSONFromBytes([]byte(heredoc.Doc(`{
		"name": "sample",
		"version": "1.0.0",
		"lockfileVersion": 1,
		"requires": true,
		"dependencies": {
			"js-tokens": {
				"version": "4.0.0",
				"resolved": "https://registry.npmjs.org/js-tokens/-/js-tokens-4.0.0.tgz",
				"integrity": "sha512-RdJUflcE3cUzKiMqQgsCu06FPu9UdIJO0beYbPhHN4k6apgJtifcoCtT9bcxOpYBtpD2kCM6Sbzg4CausW/PKQ=="
			},
			"loose-envify": {
				"version": "1.4.0",
				"resolved": "https://registry.npmjs.org/loose-envify/-/loose-envify-1.4.0.tgz",
				"integrity": "sha512-lyuxPGr/Wfhrlem2CL/UcnUc1zcqKAImBDzukY7Y5F/yQiNdko6+fRLevlw1HgMySw7f611UIY408EtxRSoK3Q==",
				"requires": {
					"js-tokens": "^3.0.0 || ^4.0.0"
				}
			}
		}
	}`)))

	req, err1 := NewAnalysisRequest(plj)
	suite.Assert().Nil(err1)

	res1, _, err1 := Packages(req, WithBaseURL(suite.server.URL))
	suite.Assert().Nil(err1)

	exp1 := &Response{
		Package{Name: "js-tokens", Version: "4.0.0", Shasum: "19203fb59991df98e3a287050d4647cdeaf32499", Verdicts: []Verdict{}},
		Package{Name: "loose-envify", Version: "1.4.0", Shasum: "71ee51fa7be4caec1a63839f7e682d8132d30caf", Verdicts: []Verdict{}},
	}
	suite.Assert().Equal(exp1, res1)

	_, res2, err2 := Packages(req, WithBaseURL(suite.server.URL), WithJSONOptions(flags.JSONFlags{JSON: true}))
	suite.Assert().Nil(err2)

	exp2 := []byte(`[{"name":"js-tokens","shasum":"19203fb59991df98e3a287050d4647cdeaf32499","verdicts":[],"version":"4.0.0"},{"name":"loose-envify","shasum":"71ee51fa7be4caec1a63839f7e682d8132d30caf","verdicts":[],"version":"1.4.0"}]`)

	suite.Assert().Equal(exp2, bytes.TrimSuffix(res2, []byte("\n")))
}

func (suite *RequestsSuite) TestVerdictsRequest() {
	req, err1 := NewVerdictsRequest([]string{"js-tokens", "4.0.0", "19203fb59991df98e3a287050d4647cdeaf32499"})
	suite.Assert().Nil(err1)

	res1, _, err1 := Packages(req, WithBaseURL(suite.server.URL))
	suite.Assert().Nil(err1)

	exp1 := &Response{
		Package{Name: "js-tokens", Version: "4.0.0", Shasum: "19203fb59991df98e3a287050d4647cdeaf32499", Verdicts: []Verdict{}},
	}
	suite.Assert().Equal(exp1, res1)

	_, res2, err2 := Packages(req, WithBaseURL(suite.server.URL), WithJSONOptions(flags.JSONFlags{JSON: true}))
	suite.Assert().Nil(err2)

	exp2 := []byte(`[{"name":"js-tokens","shasum":"19203fb59991df98e3a287050d4647cdeaf32499","verdicts":[],"version":"4.0.0"}]`)

	suite.Assert().Equal(exp2, bytes.TrimSuffix(res2, []byte("\n")))
}

func TestRequestsSuite(t *testing.T) {
	suite.Run(t, new(RequestsSuite))
}
