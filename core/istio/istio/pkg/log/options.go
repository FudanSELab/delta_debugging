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

package log

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	defaultOutputLevel        = InfoLevel
	defaultStackTraceLevel    = NoneLevel
	defaultOutputPath         = "stdout"
	defaultErrorOutputPath    = "stderr"
	defaultRotationMaxAge     = 30
	defaultRotationMaxSize    = 100 * 1024 * 1024
	defaultRotationMaxBackups = 1000
)

// Level is an enumeration of all supported log levels.
type Level string

const (
	// DebugLevel enables debug level logging
	DebugLevel Level = "debug"
	// InfoLevel enables info level logging
	InfoLevel Level = "info"
	// WarnLevel enables warn level logging
	WarnLevel Level = "warn"
	// ErrorLevel enables error level logging
	ErrorLevel Level = "error"
	// NoneLevel disables logging
	NoneLevel Level = "none"
)

// Options defines the set of options supported by Istio's component logging package.
type Options struct {
	// OutputPaths is a list of file system paths to write the log data to.
	// The special values stdout and stderr can be used to output to the
	// standard I/O streams. This defaults to stdout.
	OutputPaths []string

	// ErrorOutputPaths is a list of file system paths to write logger errors to.
	// The special values stdout and stderr can be used to output to the
	// standard I/O streams. This defaults to stderr.
	ErrorOutputPaths []string

	// RotateOutputPath is the path to a rotating log file. This file should
	// be automatically rotated over time, based on the rotation parameters such
	// as RotationMaxSize and RotationMaxAge. The default is to not rotate.
	//
	// This path is used as a foundational path. This is where log output is normally
	// saved. When a rotation needs to take place because the file got too big or too
	// old, then the file is renamed by appending a timestamp to the name. Such renamed
	// files are called backups. Once a backup has been created,
	// output resumes to this path.
	RotateOutputPath string

	// RotationMaxSize is the maximum size in megabytes of a log file before it gets
	// rotated. It defaults to 100 megabytes.
	RotationMaxSize int

	// RotationMaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename. Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is to remove log files
	// older than 30 days.
	RotationMaxAge int

	// RotationMaxBackups is the maximum number of old log files to retain.  The default
	// is to retain at most 1000 logs.
	RotationMaxBackups int

	// JSONEncoding controls whether the log is formatted as JSON.
	JSONEncoding bool

	// IncludeCallerSourceLocation determines whether log messages include the source location of the caller.
	IncludeCallerSourceLocation bool

	// LogGrpc indicates that Grpc logs should be captured. The default is true.
	// This is not exposed through the command-line flags, as this flag is mainly useful for testing: Grpc
	// stack will hold on to the logger even though it gets closed. This causes data races.
	LogGrpc bool

	stackTraceLevel string
	outputLevel     string
}

// DefaultOptions returns a new set of options, initialized to the defaults
func DefaultOptions() *Options {
	return &Options{
		OutputPaths:        []string{defaultOutputPath},
		ErrorOutputPaths:   []string{defaultErrorOutputPath},
		RotationMaxSize:    defaultRotationMaxSize,
		RotationMaxAge:     defaultRotationMaxAge,
		RotationMaxBackups: defaultRotationMaxBackups,
		outputLevel:        string(defaultOutputLevel),
		stackTraceLevel:    string(defaultStackTraceLevel),
		LogGrpc:            true,
	}
}

func isValid(level Level) bool {
	switch level {
	case DebugLevel, InfoLevel, WarnLevel, ErrorLevel, NoneLevel:
		return true
	default:
		return false
	}
}

// SetOutputLevel sets the minimum log output level.
func (o *Options) SetOutputLevel(level Level) error {
	if ok := isValid(level); !ok {
		return fmt.Errorf("unknown output level: %v", level)
	}
	o.outputLevel = string(level)
	return nil
}

// GetOutputLevel returns the minimum log output level.
func (o *Options) GetOutputLevel() (Level, error) {
	l := Level(o.outputLevel)
	if ok := isValid(l); !ok {
		return "", fmt.Errorf("unknown output level: %v", l)
	}
	return l, nil
}

// SetStackTraceLevel sets the minimum stack trace capture level.
func (o *Options) SetStackTraceLevel(level Level) error {
	if ok := isValid(level); !ok {
		return fmt.Errorf("unknown stack trace level: %v", level)
	}
	o.stackTraceLevel = string(level)
	return nil
}

// GetStackTraceLevel returns the current stack trace level.
func (o *Options) GetStackTraceLevel() (Level, error) {
	l := Level(o.stackTraceLevel)
	if ok := isValid(l); !ok {
		return "", fmt.Errorf("unknown stack trace level: %v", l)
	}
	return l, nil
}

// AttachCobraFlags attaches a set of Cobra flags to the given Cobra command.
//
// Cobra is the command-line processor that Istio uses. This command attaches
// the necessary set of flags to expose a CLI to let the user control all
// logging options.
func (o *Options) AttachCobraFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayVar(&o.OutputPaths, "log_target", o.OutputPaths,
		"The set of paths where to output the log. This can be any path as well as the special values stdout and stderr")

	cmd.PersistentFlags().StringVar(&o.RotateOutputPath, "log_rotate", o.RotateOutputPath,
		"The path for the optional rotating log file")

	cmd.PersistentFlags().IntVar(&o.RotationMaxAge, "log_rotate_max_age", o.RotationMaxAge,
		"The maximum age in days of a log file beyond which the file is rotated (0 indicates no limit)")

	cmd.PersistentFlags().IntVar(&o.RotationMaxSize, "log_rotate_max_size", o.RotationMaxSize,
		"The maximum size in megabytes of a log file beyond which the file is rotated")

	cmd.PersistentFlags().IntVar(&o.RotationMaxBackups, "log_rotate_max_backups", o.RotationMaxBackups,
		"The maximum number of log file backups to keep before older files are deleted (0 indicates no limit)")

	cmd.PersistentFlags().BoolVar(&o.JSONEncoding, "log_as_json", o.JSONEncoding,
		"Whether to format output as JSON or in plain console-friendly format")

	cmd.PersistentFlags().StringVar(&o.outputLevel, "log_output_level", o.outputLevel,
		fmt.Sprintf("The minimum logging level of messages to output, can be one of %q, %q, %q, %q, or %q",
			DebugLevel, InfoLevel, WarnLevel, ErrorLevel, NoneLevel))

	cmd.PersistentFlags().BoolVar(&o.IncludeCallerSourceLocation, "log_callers", o.IncludeCallerSourceLocation,
		"Include caller information, useful for debugging")

	cmd.PersistentFlags().StringVar(&o.stackTraceLevel, "log_stacktrace_level", o.stackTraceLevel,
		fmt.Sprintf("The minimum logging level at which stack traces are captured, can be one of %q, %q, %q, %q, or %q",
			DebugLevel, InfoLevel, WarnLevel, ErrorLevel, NoneLevel))

	// NOTE: we don't currently expose a command-line option to control ErrorOutputPaths since it
	// seems too esoteric.

	// TODO: Remove all this stuff by the 1.0 release
	var dummy string
	var dummy2 bool
	cmd.PersistentFlags().StringVarP(&dummy, "v", "v", "", "deprecated")
	cmd.PersistentFlags().StringVarP(&dummy, "stderrthreshold", "", "", "deprecated")
	cmd.PersistentFlags().BoolVarP(&dummy2, "alsologtostderr", "", false, "deprecated")
	cmd.PersistentFlags().BoolVarP(&dummy2, "logtostderr", "", false, "deprecated")

	_ = cmd.PersistentFlags().MarkDeprecated("v", "please use --log_output_level instead")
	_ = cmd.PersistentFlags().MarkShorthandDeprecated("v", "please use --log_output_level instead")
	_ = cmd.PersistentFlags().MarkDeprecated("stderrthreshold", "please use --log_output_level instead")
	_ = cmd.PersistentFlags().MarkDeprecated("logtostderr", "please use --log_target instead")
	_ = cmd.PersistentFlags().MarkDeprecated("alsologtostderr", "please use --log_target instead")
}
