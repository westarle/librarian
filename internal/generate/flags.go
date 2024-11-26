// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generate

import (
	"flag"
	"fmt"
)

type config struct {
	api      string
	language string
}

func parseFlags(cfg *config, args []string) (*config, error) {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.StringVar(&cfg.api, "api", "", "name of API inside googleapis")
	flags.StringVar(&cfg.language, "language", "", "specify from cpp, csharp, go, java, node, php, python, ruby, rust")

	// We don't want to print the whole usage message on each flags
	// error, so we set to a no-op and do the printing ourselves.
	flags.Usage = func() {}
	usage := func() {
		fmt.Fprint(flags.Output(), `Generator generates client libraries for Google APIs.

Usage:

  generator [flags]

Flags:

`)
		flags.PrintDefaults()
		fmt.Fprintf(flags.Output(), "\n\n")
	}

	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			usage() // print usage only on help
		}
		return nil, err
	}
	if err := validateConfig(cfg); err != nil {
		usage() // print usage only on help
		return nil, err
	}
	return cfg, nil
}

func validateConfig(cfg *config) error {
	if cfg.api == "" {
		return fmt.Errorf("api must be provided")
	}

	switch cfg.language {
	case "cpp":
		return errNotImplemented
	case "dotnet":
		return nil
	case "go":
		return errNotImplemented
	case "java":
		return errNotImplemented
	case "node":
		return errNotImplemented
	case "php":
		return errNotImplemented
	case "python":
		return errNotImplemented
	case "ruby":
		return errNotImplemented
	case "rust":
		return errNotImplemented
	default:
		return fmt.Errorf("invalid -language flag specified: %q", cfg.language)
	}
}
