package packagesprinter

import (
	"bytes"
	"testing"

	"github.com/cli/cli/pkg/iostreams"
)

func TestTablePrinter_printVerdictMetadata(t *testing.T) {
	tests := []struct {
		name         string
		metadata     map[string]interface{}
		wantedOutput string
	}{
		{
			name:         "empty metadata does not print anything",
			metadata:     map[string]interface{}{},
			wantedOutput: "",
		},
		{
			name: "metadata with value nil does not print anything",
			metadata: map[string]interface{}{
				"key": nil,
			},
			wantedOutput: "",
		},
		{
			name: "metadata with value prints the key and value",
			metadata: map[string]interface{}{
				"key": "myvalue",
			},
			wantedOutput: "    key: myvalue\n",
		},
		{
			name: "metadata with value prints the key and value while ignoring npm_package_name and npm_package_version",
			metadata: map[string]interface{}{
				"key":                 "myvalue",
				"npm_package_name":    "react",
				"npm_package_version": "0.18.0",
			},
			wantedOutput: "    key: myvalue\n",
		},
		{
			name: "metadata with values prints the values it recognizes and ignores the rest",
			metadata: map[string]interface{}{
				"mystringkey":   "a string",
				"myintkey":      10,
				"somethingelse": float64(10.20),
			},
			wantedOutput: "    myintkey: 10\n    mystringkey: a string\n",
		},
		{
			name: "metadata with value prints the key and value while ignoring empty values",
			metadata: map[string]interface{}{
				"parent_name:":    "node",
				"executable_path": "/bin/sh",
				"commandline":     `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
				"server_ip":       "",
			},
			wantedOutput: "    commandline: sh -c  node -e \"try{require('./_postinstall')}catch(e){}\" || exit 0\n    executable_path: /bin/sh\n    parent_name:: node\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}
			tr := &TablePrinter{

				streams: &iostreams.IOStreams{
					Out:    outBuf,
					ErrOut: errBuf,
				},
			}
			tr.printVerdictMetadata(tt.metadata)

			if tt.wantedOutput != outBuf.String() {
				t.Errorf("wanted output %q, got %q", tt.wantedOutput, outBuf.String())
			}
		})
	}
}
