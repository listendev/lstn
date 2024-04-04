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
	"context"
	"encoding/json"
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
	"github.com/listendev/lstn/pkg/pypi"
	"github.com/listendev/pkg/ecosystem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	localEndpoint    = "http://127.0.0.1:3000"
	nonLocalEndpoint = "https://smtg.listen.dev"
)

type mockContextLocalEndpoint struct{}

func strPtr(s string) *string {
	return &s
}

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
	if key == pkgcontext.ConfigKey {
		c, _ := flags.NewConfigFlags()
		c.Endpoint.Npm = localEndpoint

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
	require.Nil(t, e)

	if _, err := getEndpointURLFromContext(&AnalysisRequest{}, o); assert.Error(t, err) {
		assert.Equal(t, "couldn't obtain configuration options", err.Error())
	}

	if _, err := getEndpointURLFromContext(&VerdictsRequest{}, o); assert.Error(t, err) {
		assert.Equal(t, "couldn't obtain configuration options", err.Error())
	}

	o1, e1 := newOptions(WithContext(mockContextLocalEndpoint{}))
	require.Nil(t, e1)
	if _, err := getEndpointURLFromContext(&AnalysisRequest{}, o1); assert.Error(t, err) {
		assert.Equal(t, "couldn't obtain ecosystem from options", err.Error())
	}

	if _, err := getEndpointURLFromContext(&VerdictsRequest{}, o1); assert.Error(t, err) {
		assert.Equal(t, "couldn't obtain ecosystem from options", err.Error())
	}
}

func TestLocalEndpoint(t *testing.T) {
	eco := ecosystem.Npm
	o, e := newOptions(WithContext(mockContextLocalEndpoint{}), WithEcosystem(eco))
	assert.Nil(t, e)

	endpointAnalysis, err := getEndpointURLFromContext(&AnalysisRequest{}, o)
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(endpointAnalysis, localEndpoint))
	pAnalysis, eAnalysis := getAPIPrefix(endpointAnalysis, eco)
	require.Nil(t, eAnalysis)
	assert.Equal(t, "/api/npm", pAnalysis)
	assert.True(t, strings.HasSuffix(endpointAnalysis, "analysis"))

	endpointVerdicts, err := getEndpointURLFromContext(&VerdictsRequest{}, o)
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(endpointVerdicts, localEndpoint))
	pVerdicts, eVerdicts := getAPIPrefix(endpointVerdicts, eco)
	require.Nil(t, eVerdicts)
	assert.Equal(t, "/api/npm", pVerdicts)
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
		c.Endpoint.Npm = nonLocalEndpoint

		return c
	}

	return nil
}

func TestNonLocalEndpoint(t *testing.T) {
	eco := ecosystem.Npm
	o, e := newOptions(WithContext(mockContextNonLocalEndpoint{}), WithEcosystem(eco))
	assert.Nil(t, e)

	endpointAnalysis, err := getEndpointURLFromContext(&AnalysisRequest{}, o)
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(endpointAnalysis, nonLocalEndpoint))
	pAnalysis, eAnalysis := getAPIPrefix(endpointAnalysis, eco)
	require.Nil(t, eAnalysis)
	assert.Equal(t, "/api", pAnalysis)
	assert.True(t, strings.HasSuffix(endpointAnalysis, "analysis"))

	endpointVerdicts, err := getEndpointURLFromContext(&VerdictsRequest{}, o)
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(endpointVerdicts, nonLocalEndpoint))
	pVerdicts, eVerdicts := getAPIPrefix(endpointVerdicts, eco)
	require.Nil(t, eVerdicts)
	assert.Equal(t, "/api", pVerdicts)
	assert.True(t, strings.HasSuffix(endpointVerdicts, "verdicts"))
}

type RequestsSuite struct {
	suite.Suite
	server *httptest.Server
}

func (suite *RequestsSuite) BeforeTest(suiteName, testName string) {
	suite.Assert().Equal("RequestsSuite", suiteName)

	switch testName {
	case "TestNPMAnalysisRequest":
		resp := []byte(`[{"name":"js-tokens","digest":"19203fb59991df98e3a287050d4647cdeaf32499","verdicts":[],"version":"4.0.0"},{"name":"loose-envify","digest":"71ee51fa7be4caec1a63839f7e682d8132d30caf","verdicts":[],"version":"1.4.0"}]`)
		suite.server = internaltesting.MockHTTPServer(suite.Assert(), "analysis", resp, http.StatusOK, "POST")
	case "TestNPMVerdictsRequest":
		resp := []byte(`[{"name":"js-tokens","digest":"19203fb59991df98e3a287050d4647cdeaf32499","verdicts":[],"version":"4.0.0"}]`)
		suite.server = internaltesting.MockHTTPServer(suite.Assert(), "verdicts", resp, http.StatusOK, "POST")
	case "TestPyPiAnalysisRequest":
		resp := []byte(`[{"name":"click","digest":"598784326af34517fca8c58418d148f2403df25303e02736832403587318e9e8","verdicts":[],"version":"8.1.3"},{"name":"colorama","digest":"d8536f443c9a4a8358a93a6792e2acffb9d9d5cb0a5cfd8802644b7b1c9a02e4","verdicts":[],"version":"0.4.6"}]`)
		suite.server = internaltesting.MockHTTPServer(suite.Assert(), "analysis", resp, http.StatusOK, "POST")
	case "TestPyPiVerdictsRequest":
		resp := []byte(`[{"name":"colorama","digest":"d8536f443c9a4a8358a93a6792e2acffb9d9d5cb0a5cfd8802644b7b1c9a02e4","verdicts":[],"version":"0.4.6"}]`)
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
	resp, data, erro := Packages(&AnalysisRequest{}, WithContext(mockContextLocalEndpoint{}), WithEcosystem(ecosystem.Npm))
	suite.Assert().Nil(resp)
	suite.Assert().Nil(data)
	if suite.Assert().Error(erro) {
		suite.Assert().Contains(erro.Error(), "manifest is mandatory")
	}
}

func (suite *RequestsSuite) TestVerdictsRequestEmpty() {
	resp, data, erro := Packages(&VerdictsRequest{}, WithContext(mockContextLocalEndpoint{}), WithEcosystem(ecosystem.Npm))
	suite.Assert().Nil(resp)
	suite.Assert().Nil(data)
	if suite.Assert().Error(erro) {
		suite.Assert().Contains(erro.Error(), "name is mandatory")
	}
}

func (suite *RequestsSuite) TestNPMAnalysisRequest() {
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

	req, err1 := NewAnalysisRequest(plj, WithRequestContext())
	suite.Assert().Nil(err1)

	res1, _, err1 := Packages(req, WithBaseURL(suite.server.URL), WithEcosystem(ecosystem.Npm))
	suite.Assert().Nil(err1)

	exp1 := &Response{
		Package{Name: "js-tokens", Version: strPtr("4.0.0"), Digest: strPtr("19203fb59991df98e3a287050d4647cdeaf32499"), Verdicts: []Verdict{}},
		Package{Name: "loose-envify", Version: strPtr("1.4.0"), Digest: strPtr("71ee51fa7be4caec1a63839f7e682d8132d30caf"), Verdicts: []Verdict{}},
	}
	suite.Assert().Equal(exp1, res1)

	_, res2, err2 := Packages(req, WithBaseURL(suite.server.URL), WithJSONOptions(flags.JSONFlags{JSON: true}), WithEcosystem(ecosystem.Npm))
	suite.Assert().Nil(err2)

	exp2 := &Response{}
	suite.Assert().Nil(json.Unmarshal(res2, exp2))
	suite.Assert().Equal(exp2, exp1)
}

func (suite *RequestsSuite) TestPyPiAnalysisRequest() {
	plj, _ := pypi.NewPoetryLockFromBytes([]byte(heredoc.Doc(`# This file is automatically @generated by Poetry and should not be changed by hand.

[[package]]
name = "click"
version = "8.1.3"
description = "Composable command line interface toolkit"
category = "main"
optional = false
python-versions = ">=3.7"
files = [
	{file = "click-8.1.3-py3-none-any.whl", hash = "sha256:bb4d8133cb15a609f44e8213d9b391b0809795062913b383c62be0ee95b1db48"},
	{file = "click-8.1.3.tar.gz", hash = "sha256:7682dc8afb30297001674575ea00d1814d808d6a36af415a82bd481d37ba7b8e"},
]

[package.dependencies]
colorama = {version = "*", markers = "platform_system == \"Windows\""}

[[package]]
name = "colorama"
version = "0.4.6"
description = "Cross-platform colored terminal text."
category = "main"
optional = false
python-versions = "!=3.0.*,!=3.1.*,!=3.2.*,!=3.3.*,!=3.4.*,!=3.5.*,!=3.6.*,>=2.7"
files = [
	{file = "colorama-0.4.6-py2.py3-none-any.whl", hash = "sha256:4f1d9991f5acc0ca119f9d443620b77f9d6b33703e51011c16baf57afb285fc6"},
	{file = "colorama-0.4.6.tar.gz", hash = "sha256:08695f5cb7ed6e0531a20572697297273c47b8cae5a63ffc6d6ed5c201be6e44"},
]`)))

	req, err1 := NewAnalysisRequest(plj, WithRequestContext())
	suite.Assert().Nil(err1)

	res1, _, err1 := Packages(req, WithBaseURL(suite.server.URL), WithEcosystem(ecosystem.Pypi))
	suite.Assert().Nil(err1)

	exp1 := &Response{
		Package{Name: "click", Version: strPtr("8.1.3"), Digest: strPtr("598784326af34517fca8c58418d148f2403df25303e02736832403587318e9e8"), Verdicts: []Verdict{}},
		Package{Name: "colorama", Version: strPtr("0.4.6"), Digest: strPtr("d8536f443c9a4a8358a93a6792e2acffb9d9d5cb0a5cfd8802644b7b1c9a02e4"), Verdicts: []Verdict{}},
	}
	suite.Assert().Equal(exp1, res1)

	_, res2, err2 := Packages(req, WithBaseURL(suite.server.URL), WithJSONOptions(flags.JSONFlags{JSON: true}), WithEcosystem(ecosystem.Pypi))
	suite.Assert().Nil(err2)

	exp2 := &Response{}
	suite.Assert().Nil(json.Unmarshal(res2, exp2))
	suite.Assert().Equal(exp2, exp1)
}

func (suite *RequestsSuite) TestNPMVerdictsRequest() {
	req, err1 := NewVerdictsRequest([]string{"js-tokens", "4.0.0", "19203fb59991df98e3a287050d4647cdeaf32499"})
	suite.Assert().Nil(err1)

	res1, _, err1 := Packages(req, WithBaseURL(suite.server.URL), WithEcosystem(ecosystem.Npm))
	suite.Assert().Nil(err1)

	exp1 := &Response{
		Package{Name: "js-tokens", Version: strPtr("4.0.0"), Digest: strPtr("19203fb59991df98e3a287050d4647cdeaf32499"), Verdicts: []Verdict{}},
	}
	suite.Assert().Equal(exp1, res1)

	_, res2, err2 := Packages(req, WithBaseURL(suite.server.URL), WithJSONOptions(flags.JSONFlags{JSON: true}), WithEcosystem(ecosystem.Npm))
	suite.Assert().Nil(err2)

	exp2 := &Response{}
	suite.Assert().Nil(json.Unmarshal(res2, exp2))
	suite.Assert().Equal(exp1, exp2)
}

func (suite *RequestsSuite) TestPyPiVerdictsRequest() {
	req, err1 := NewVerdictsRequest([]string{"colorama", "0.4.6", "d8536f443c9a4a8358a93a6792e2acffb9d9d5cb0a5cfd8802644b7b1c9a02e4"})
	suite.Assert().Nil(err1)

	res1, _, err1 := Packages(req, WithBaseURL(suite.server.URL), WithEcosystem(ecosystem.Pypi))
	suite.Assert().Nil(err1)

	exp1 := &Response{
		Package{Name: "colorama", Version: strPtr("0.4.6"), Digest: strPtr("d8536f443c9a4a8358a93a6792e2acffb9d9d5cb0a5cfd8802644b7b1c9a02e4"), Verdicts: []Verdict{}},
	}
	suite.Assert().Equal(exp1, res1)

	_, res2, err2 := Packages(req, WithBaseURL(suite.server.URL), WithJSONOptions(flags.JSONFlags{JSON: true}), WithEcosystem(ecosystem.Pypi))
	suite.Assert().Nil(err2)

	exp2 := &Response{}
	suite.Assert().Nil(json.Unmarshal(res2, exp2))
	suite.Assert().Equal(exp1, exp2)
}

func TestRequestsSuite(t *testing.T) {
	suite.Run(t, new(RequestsSuite))
}
