/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This generator is based on Kubernetes' prerelease-lifecycle-gen
// but is being used to generate files to be consumed externally (like Kubepug, but also
// eventually to generate a better deprecations page)

package main

import (
	"encoding/json"
	"flag"
	"fmt"

	deprecationsgenerator "github.com/kubepug/kubepug/generator/deprecations"

	"github.com/spf13/pflag"
	"k8s.io/code-generator/cmd/prerelease-lifecycle-gen/args"
	"k8s.io/gengo/v2"
	"k8s.io/gengo/v2/generator"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)

	argsd := args.New()

	argsd.AddFlags(pflag.CommandLine)
	flag.Set("logtostderr", "true")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	if err := argsd.Validate(); err != nil {
		klog.Fatalf("Error: %v", err)
	}

	regGenerator := deprecationsgenerator.NewAPIRegistry()
	myTargets := func(context *generator.Context) []generator.Target {
		return regGenerator.GetTargets(context, argsd)
	}

	if err := gengo.Execute(
		deprecationsgenerator.NameSystems(),
		deprecationsgenerator.DefaultNameSystem(),
		myTargets,
		gengo.StdBuildTag,
		pflag.Args(),
	); err != nil {
		klog.Errorf("error generating some files, may have missing status: %s", err)
	}

	registries := regGenerator.Registry()
	data, err := json.Marshal(registries)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
}
