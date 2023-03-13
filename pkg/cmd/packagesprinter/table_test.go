package packagesprinter

import (
	"bytes"
	"testing"

	"github.com/cli/cli/pkg/iostreams"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/stretchr/testify/require"
)

func TestTablePrinter_printVerdictMetadata(t *testing.T) {
	tests := []struct {
		name           string
		metadata       map[string]interface{}
		expectedOutput string
	}{
		{
			name:           "empty metadata does not print anything",
			metadata:       map[string]interface{}{},
			expectedOutput: "",
		},
		{
			name: "metadata with value nil does not print anything",
			metadata: map[string]interface{}{
				"key": nil,
			},
			expectedOutput: "",
		},
		{
			name: "metadata with value prints the key and value",
			metadata: map[string]interface{}{
				"key": "myvalue",
			},
			expectedOutput: "    key: myvalue\n",
		},
		{
			name: "metadata with value prints the key and value while ignoring npm_package_name and npm_package_version",
			metadata: map[string]interface{}{
				"key":                 "myvalue",
				"npm_package_name":    "react",
				"npm_package_version": "0.18.0",
			},
			expectedOutput: "    key: myvalue\n",
		},
		{
			name: "metadata with values prints the values it recognizes and ignores the rest",
			metadata: map[string]interface{}{
				"mystringkey":   "a string",
				"myintkey":      10,
				"somethingelse": float64(10.20),
			},
			expectedOutput: "    myintkey: 10\n    mystringkey: a string\n",
		},
		{
			name: "metadata with value prints the key and value while ignoring empty values",
			metadata: map[string]interface{}{
				"parent_name:":    "node",
				"executable_path": "/bin/sh",
				"commandline":     `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
				"server_ip":       "",
			},
			expectedOutput: "    commandline: sh -c  node -e \"try{require('./_postinstall')}catch(e){}\" || exit 0\n    executable_path: /bin/sh\n    parent_name:: node\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			tr := &TablePrinter{

				streams: &iostreams.IOStreams{
					Out: outBuf,
				},
			}
			tr.printVerdictMetadata(tt.metadata)

			require.Equal(t, tt.expectedOutput, outBuf.String())
		})
	}
}

func TestTablePrinter_printVerdict(t *testing.T) {
	tests := []struct {
		name           string
		p              *listen.Package
		verdict        listen.Verdict
		expectedOutput string
	}{
		{
			name: "verdict with metadata prints the verdict and metadata",
			p: &listen.Package{
				Name:     "my-package",
				Version:  "1.0.0",
				Verdicts: []listen.Verdict{},

				Problems: []listen.Problem{},
			},
			verdict: listen.Verdict{
				Message:  "outbound network connection",
				Priority: "high",
				Metadata: map[string]interface{}{
					"parent_name":     "node",
					"executable_path": "/bin/sh",
					"commandline":     `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
					"server_ip":       "",
				},
			},
			expectedOutput: "  [high] outbound network connection\n    commandline: sh -c  node -e \"try{require('./_postinstall')}catch(e){}\" || exit 0\n    executable_path: /bin/sh\n    parent_name: node\n",
		},
		{
			name: "verdict with transitive metadata marks the verdict as transitive",
			p: &listen.Package{
				Name:     "my-package",
				Version:  "1.0.0",
				Verdicts: []listen.Verdict{},

				Problems: []listen.Problem{},
			},
			verdict: listen.Verdict{
				Message:  "outbound network connection",
				Priority: "high",
				Metadata: map[string]interface{}{
					"npm_package_name":    "react",
					"npm_package_version": "0.18.0",
					"parent_name":         "node",
					"executable_path":     "/bin/sh",
					"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
					"server_ip":           "",
				},
			},
			expectedOutput: "  [high] outbound network connection (from transitive dependency react@0.18.0)\n    commandline: sh -c  node -e \"try{require('./_postinstall')}catch(e){}\" || exit 0\n    executable_path: /bin/sh\n    parent_name: node\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			tr := &TablePrinter{
				streams: &iostreams.IOStreams{
					Out: outBuf,
				},
			}
			tr.printVerdict(tt.p, tt.verdict)

			require.Equal(t, tt.expectedOutput, outBuf.String())
		})
	}
}

func TestTablePrinter_printProblem(t *testing.T) {
	tests := []struct {
		name           string
		problem        listen.Problem
		expectedOutput string
	}{
		{
			name: "problem with all details gets printed",
			problem: listen.Problem{
				Type:   "https://listen.dev/probs/invalid-name",
				Title:  "Package name not valid",
				Detail: "Package name not valid",
			},
			expectedOutput: "  - Package name not valid: https://listen.dev/probs/invalid-name\n",
		},
		{
			name: "problem with missing type",
			problem: listen.Problem{
				Type:   "",
				Title:  "Package name not valid",
				Detail: "Package name not valid",
			},
			expectedOutput: "  - Package name not valid: \n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			tr := &TablePrinter{
				streams: &iostreams.IOStreams{
					Out: outBuf,
				},
			}
			tr.printProblem(tt.problem)

			require.Equal(t, tt.expectedOutput, outBuf.String())
		})
	}
}
