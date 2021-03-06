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

// +build integration

package artifact

import (
	"os"
	"path/filepath"
	"testing"

	"fmt"

	"github.com/aws/amazon-ssm-agent/agent/log"
	"github.com/stretchr/testify/assert"
)

type DownloadTest struct {
	input          DownloadInput
	expectedOutput DownloadOutput
}

var (
	localPathExist, _    = filepath.Abs(filepath.Join(".", "testdata", "CheckMyHash.txt"))
	localPathNotExist, _ = filepath.Abs(filepath.Join(".", "testdata", "IDontExist.txt"))
	downloadFolder, _    = filepath.Abs(filepath.Join(".", "testdata"))
	mockLog              = log.NewMockLog()

	downloadTests = []DownloadTest{
		// {DownloadInput{SourceUrl, DestinationDirectory, SourceHashValue, SourceHashType},
		// DownloadOutput{LocalFilePath, IsUpdated, IsHashMatched}},
		{
			// validate sha256
			DownloadInput{
				localPathExist,
				downloadFolder,
				"090c1965e46155b2b23ba9093ed7c67243957a397e3ad5531a693d57958a760a",
				"sha256"},
			DownloadOutput{
				localPathExist,
				false,
				true},
		},
		{
			// validate incorrect sha256 fails
			DownloadInput{
				localPathExist,
				downloadFolder,
				"111111111",
				"sha256"},
			DownloadOutput{
				localPathExist,
				false,
				false},
		},
		{
			// validate md5
			DownloadInput{
				localPathExist,
				downloadFolder,
				"e84913ff3a8eef39238b32170e657ba8",
				"md5"},
			DownloadOutput{
				localPathExist,
				false,
				true},
		},

		{
			// validate incorrect md5 fails
			DownloadInput{
				localPathExist,
				downloadFolder,
				"222222222",
				"md5"},
			DownloadOutput{
				localPathExist,
				false,
				false},
		},
		{
			// ensure default is sha256
			DownloadInput{
				localPathExist,
				downloadFolder,
				"090c1965e46155b2b23ba9093ed7c67243957a397e3ad5531a693d57958a760a",
				""},
			DownloadOutput{
				localPathExist,
				false,
				true},
		},
		{
			// relative url is not supported
			DownloadInput{
				"IamRelativeFilePath",
				downloadFolder,
				"090c1965e46155b2b23ba9093ed7c67243957a397e3ad5531a693d57958a760a",
				""},
			DownloadOutput{
				"",
				false,
				false},
		},
		{
			// relative url is not supported
			DownloadInput{
				"IamRelativeFilePath/IdontExist",
				downloadFolder,
				"090c1965e46155b2b23ba9093ed7c67243957a397e3ad5531a693d57958a760a",
				""},
			DownloadOutput{
				"",
				false,
				false},
		},
		{
			// s3 download error
			DownloadInput{
				"https://s3.amazonaws.com/ssmnotsuchbucket/ssmnosuchfile.txt",
				downloadFolder,
				"",
				""},
			DownloadOutput{
				"",
				false,
				false},
		},
	}
)

func runDownloadTests(t *testing.T, tests []DownloadTest) {
	for _, test := range tests {
		output, err := Download(mockLog, test.input)
		t.Log(err)
		assert.Equal(t, test.expectedOutput, output)
	}
}

func TestDownloads(t *testing.T) {
	x, y := os.Getwd()
	mockLog.Infof("Working Directory is %v, %v", x, y)
	runDownloadTests(t, downloadTests)
}

func TestHttpHttpsDownloadArtifact(t *testing.T) {
	testFilePath := "https://www.ietf.org/rfc/rfc1350.txt"
	downloadInput := DownloadInput{
		DestinationDirectory: ".",
		SourceURL:            testFilePath,
		SourceHashValue:      "39c9534e5fa6fecd3ac083ffd6256c2cc9a58f9f1058cb2e472d1782040231f9",
		SourceHashType:       "sha256",
	}
	var expectedLocalPath = "23ac2f78ce2a6048417d8594fcc4eb9f_rfc1350.txt"
	os.Remove(expectedLocalPath)
	os.Remove(expectedLocalPath + ".etag")
	expectedOutput := DownloadOutput{
		expectedLocalPath,
		true,
		true}
	output, err := Download(mockLog, downloadInput)
	assert.NoError(t, err, "Failed to download %v", downloadInput)
	mockLog.Infof("Download Result is %v and err:%v", output, err)
	assert.Equal(t, expectedOutput, output)

	// now since we have downloaded the file, try to download again should result in cache hit!
	expectedOutput = DownloadOutput{
		expectedLocalPath,
		false,
		true}
	output, err = Download(mockLog, downloadInput)
	assert.NoError(t, err, "Failed to download %v", downloadInput)
	mockLog.Infof("Download Result is %v and err:%v", output, err)
	assert.Equal(t, expectedOutput, output)

	os.Remove(expectedLocalPath)
	os.Remove(expectedLocalPath + ".etag")
}

func ExampleMd5HashValue() {
	path := filepath.Join("testdata", "CheckMyHash.txt")
	mockLog := log.NewMockLog()
	content, _ := Md5HashValue(mockLog, path)
	fmt.Println(content)
	// Output:
	// e84913ff3a8eef39238b32170e657ba8
}

func ExampleSha256HashValue() {
	path := filepath.Join("testdata", "CheckMyHash.txt")
	mockLog := log.NewMockLog()
	content, _ := Sha256HashValue(mockLog, path)
	fmt.Println(content)
	// Output:
	// 090c1965e46155b2b23ba9093ed7c67243957a397e3ad5531a693d57958a760a
}
