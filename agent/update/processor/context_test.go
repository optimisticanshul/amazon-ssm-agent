// Copyright 2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Amazon Software License (the "License"). You may not
// use this file except in compliance with the License. A copy of the
// License is located at
//
// http://aws.amazon.com/asl/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Package processor contains the methods for update ssm agent.
// It also provides methods for sendReply and updateInstanceInfo
package processor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/aws/amazon-ssm-agent/agent/log"
	"github.com/stretchr/testify/assert"
)

var logger = log.NewMockLog()

//Valid context files
var sampleFiles = []string{
	"testdata/updatecontext.json",
}

type testCase struct {
	Input    string
	FileName string
	Output   UpdateContext
}

type contextMgrStub struct{}

func (c *contextMgrStub) saveUpdateContext(log log.T, context *UpdateContext, contextLocation string) (err error) {
	return nil
}

func (c *contextMgrStub) uploadOutput(log log.T, context *UpdateContext) error {
	return nil
}

type serviceStub struct {
	Service
}

func (s *serviceStub) SendReply(log log.T, update *UpdateDetail) error {
	return nil
}

func (s *serviceStub) DeleteMessage(log log.T, update *UpdateDetail) error {
	return nil
}

func (s *serviceStub) UpdateHealthCheck(log log.T, update *UpdateDetail, errorCode string) error {
	return nil
}

type FakeServiceWithFailure struct {
	Service
}

func (s *FakeServiceWithFailure) SendReply(log log.T, update *UpdateDetail) error {
	return fmt.Errorf("Failed to send reply")
}

func (s *FakeServiceWithFailure) DeleteMessage(log log.T, update *UpdateDetail) error {
	return fmt.Errorf("Failed to delete message")
}

func (s *FakeServiceWithFailure) UpdateHealthCheck(log log.T, update *UpdateDetail, errorCode string) error {
	return fmt.Errorf("Failed to update health check")
}

//Test loadContextFromFile file with valid context files
func TestParseContext(t *testing.T) {
	// generate test cases
	var testCases []testCase
	for _, contextFile := range sampleFiles {
		testCases = append(testCases, testCase{
			Input:    string(loadFile(t, contextFile)),
			Output:   loadContextFromFile(t, contextFile),
			FileName: contextFile,
		})
	}

	// run tests
	for _, tst := range testCases {
		// call method
		parsedContext, err := parseContext(log.GetLogger(), tst.FileName)

		// check results
		assert.Nil(t, err)
		assert.Equal(t, parsedContext.Current.SourceVersion, "")
		assert.Equal(t, len(parsedContext.Histories), 1)
		assert.Equal(t, parsedContext.Histories[0].SourceVersion, "0.0.3.0")
	}
}

type ContextTestCase struct {
	Context      *UpdateContext
	InfoMessage  string
	ErrorMessage string
	HasMessageID bool
}

var contextTests = []ContextTestCase{
	generateTestCase(),
}

func TestContext(t *testing.T) {
	// run tests
	for _, tst := range contextTests {
		// call method
		hasMessageID := tst.Context.Current.HasMessageID()
		// check results
		assert.Equal(t, hasMessageID, tst.HasMessageID)

		tst.Context.Current.AppendInfo(logger, tst.InfoMessage)
		assert.Equal(t, tst.Context.Current.StandardOut, tst.InfoMessage)

		tst.Context.Current.AppendError(logger, tst.ErrorMessage)
		assert.Equal(t, tst.Context.Current.StandardOut, fmt.Sprintf("%v\n%v", tst.InfoMessage, tst.ErrorMessage))
		assert.Equal(t, tst.Context.Current.StandardError, tst.ErrorMessage)
	}
}

func TestIsUpdateInProgress(t *testing.T) {
	context := generateTestCase().Context
	context.Current.StartDateTime = time.Now()
	context.Current.State = Staged

	result := context.IsUpdateInProgress(logger)

	assert.True(t, result)
}

func TestLoadUpdateContext(t *testing.T) {
	context, err := LoadUpdateContext(logger, sampleFiles[0])

	assert.NoError(t, err)
	assert.True(t, len(context.Histories) > 0)
}

func generateTestCase() ContextTestCase {
	testCase := ContextTestCase{
		Context:      &UpdateContext{},
		InfoMessage:  "Test Message",
		ErrorMessage: "Error Message",
		HasMessageID: true,
	}

	testCase.Context.Current = &UpdateDetail{
		MessageID: "MessageId",
	}
	testCase.Context.Histories = []*UpdateDetail{}
	return testCase
}

//Load specified file from file system
func loadFile(t *testing.T, fileName string) (result []byte) {
	result, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return
}

//Parse manifest file
func loadContextFromFile(t *testing.T, fileName string) (context UpdateContext) {
	b := loadFile(t, fileName)
	err := json.Unmarshal(b, &context)
	if err != nil {
		t.Fatal(err)
	}
	return context
}
