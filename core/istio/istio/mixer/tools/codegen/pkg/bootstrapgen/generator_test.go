// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bootstrapgen

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"
)

type logFn func(string, ...interface{})

// TestGenerator_Generate uses the outputs file descriptors generated via bazel
// and compares them against the golden files.
func TestGenerator_Generate(t *testing.T) {
	importmap := map[string]string{
		"mixer/v1/config/descriptor/value_type.proto": "istio.io/api/mixer/v1/config/descriptor",
		"mixer/v1/template/extensions.proto":          "istio.io/api/mixer/v1/template",
		"gogoproto/gogo.proto":                        "github.com/gogo/protobuf/gogoproto",
		"google/protobuf/duration.proto":              "github.com/gogo/protobuf/types",
	}

	tests := []struct {
		name     string
		fdsFiles map[string]string // FDS and their package import paths
		want     string
	}{
		{"AllTemplates", map[string]string{
			"testdata/check/template_proto.descriptor_set":   "istio.io/istio/mixer/template/list",
			"testdata/report2/template_proto.descriptor_set": "istio.io/istio/mixer/template/metric",
			"testdata/quota/template_proto.descriptor_set":   "istio.io/istio/mixer/template/quota",
			"testdata/apa/template_proto.descriptor_set":     "istio.io/istio/mixer/template/apa",
			"testdata/report1/template_proto.descriptor_set": "istio.io/istio/mixer/template/log"},
			"testdata/template.gen.go.golden"},
	}
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			testTmpDir := path.Join(os.TempDir(), "bootstrapTemplateTest")
			_ = os.MkdirAll(testTmpDir, os.ModeDir|os.ModePerm)
			outFile, err := os.Create(path.Join(testTmpDir, path.Base(v.want)))
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if !t.Failed() {
					if removeErr := os.RemoveAll(testTmpDir); removeErr != nil {
						t.Logf("Could not remove temporary folder %s: %v", testTmpDir, removeErr)

					}
				} else {
					t.Logf("Generated data is located at '%s'", testTmpDir)
				}
			}()

			g := Generator{OutFilePath: outFile.Name(), ImportMapping: importmap}
			if err := g.Generate(v.fdsFiles); err != nil {
				t.Fatalf("Generate(%s) produced an error: %v", v.fdsFiles, err)
			}

			if same := fileCompare(outFile.Name(), v.want, t.Errorf); !same {
				t.Error("Files were not the same.")
			}
		})
	}
}

const chunkSize = 64000

func fileCompare(file1, file2 string, logf logFn) bool {
	f1, err := os.Open(file1)
	if err != nil {
		logf("could not open file: %v", err)
		return false
	}

	f2, err := os.Open(file2)
	if err != nil {
		logf("could not open file: %v", err)
		return false
	}

	for {
		b1 := make([]byte, chunkSize)
		s1, err1 := f1.Read(b1)

		b2 := make([]byte, chunkSize)
		s2, err2 := f2.Read(b2)

		if err1 == io.EOF && err2 == io.EOF {
			return true
		}

		if err1 != nil || err2 != nil {
			return false
		}

		if !bytes.Equal(b1, b2) {
			logf("bytes don't match (sizes: %d, %d):\n%s\n%s", s1, s2, string(b1), string(b2))
			return false
		}
	}
}
