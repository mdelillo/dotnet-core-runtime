package main

import (
	"github.com/BurntSushi/toml"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dotnet-core-runtime-cnb/runtime"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, _ spec.G, it spec.S) {
	var factory *test.DetectFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewDetectFactory(t)
		fakeBuildpackToml := `
[[dependencies]]
id = "dotnet-runtime"
name = "Dotnet Runtime"
stacks = ["org.testeroni"]
uri = "some-uri"
version = "2.2.5"
`
		_, err := toml.Decode(fakeBuildpackToml, &factory.Detect.Buildpack.Metadata)
		Expect(err).ToNot(HaveOccurred())
		factory.Detect.Stack = "org.testeroni"

	})

	it("passes when there is a valid runtimeconfig.json where the specified version is provided", func() {
		runtimeConfigJSONPath := filepath.Join(factory.Detect.Application.Root, "appName.runtimeconfig.json")
		Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
{
  "runtimeOptions": {
    "tfm": "netcoreapp2.2",
    "framework": {
      "name": "Microsoft.NETCore.App",
      "version": "2.2.5"
    }
  }
}
`), os.ModePerm)).To(Succeed())
		code, err := runDetect(factory.Detect)
		Expect(err).ToNot(HaveOccurred())
		Expect(code).To(Equal(detect.PassStatusCode))
		Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
			Provides: []buildplan.Provided{{Name: runtime.DotnetRuntime}},
			Requires: []buildplan.Required{{
				Name: runtime.DotnetRuntime,
				Version: "2.2.5",
				Metadata: buildplan.Metadata{"launch": true},
			}},
		}))
	})

	it("passes when there is a valid runtimeconfig.json where the specified minor is provided", func() {
		runtimeConfigJSONPath := filepath.Join(factory.Detect.Application.Root, "appName.runtimeconfig.json")
		Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
{
  "runtimeOptions": {
    "tfm": "netcoreapp2.2",
    "framework": {
      "name": "Microsoft.NETCore.App",
      "version": "2.2.0"
    }
  }
}
`), os.ModePerm)).To(Succeed())
		code, err := runDetect(factory.Detect)
		Expect(err).ToNot(HaveOccurred())
		Expect(code).To(Equal(detect.PassStatusCode))
		Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
			Provides: []buildplan.Provided{{Name: runtime.DotnetRuntime}},
			Requires: []buildplan.Required{{
				Name: runtime.DotnetRuntime,
				Version: "2.2.5",
				Metadata: buildplan.Metadata{"launch": true},
			}},
		}))
	})

	it("passes when there is a valid runtimeconfig.json where the specified major is provided", func() {
		runtimeConfigJSONPath := filepath.Join(factory.Detect.Application.Root, "appName.runtimeconfig.json")
		Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
{
  "runtimeOptions": {
    "tfm": "netcoreapp2.2",
    "framework": {
      "name": "Microsoft.NETCore.App",
      "version": "2.1.0"
    }
  }
}
`), os.ModePerm)).To(Succeed())
		code, err := runDetect(factory.Detect)
		Expect(err).ToNot(HaveOccurred())
		Expect(code).To(Equal(detect.PassStatusCode))
		Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
			Provides: []buildplan.Provided{{Name: runtime.DotnetRuntime}},
			Requires: []buildplan.Required{{
				Name: runtime.DotnetRuntime,
				Version: "2.2.5",
				Metadata: buildplan.Metadata{"launch": true},
			}},
		}))
	})

	it("passes when there is a valid runtimeconfig.json where there are no valid roll forward versions available", func() {
		runtimeConfigJSONPath := filepath.Join(factory.Detect.Application.Root, "appName.runtimeconfig.json")
		Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
{
  "runtimeOptions": {
    "tfm": "netcoreapp2.2",
    "framework": {
      "name": "Microsoft.NETCore.App",
      "version": "1.1.0"
    }
  }
}
`), os.ModePerm)).To(Succeed())
		code, err := runDetect(factory.Detect)
		Expect(err).To(HaveOccurred())
		Expect(code).To(Equal(detect.FailStatusCode))
	})

	it("passes when there is a valid runtimeconfig.json where there are no runtime options meaning the app is a self contained deployment", func() {
		runtimeConfigJSONPath := filepath.Join(factory.Detect.Application.Root, "appName.runtimeconfig.json")
		Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
{
 "runtimeOptions": {}
}
`), os.ModePerm)).To(Succeed())
		code, err := runDetect(factory.Detect)
		Expect(err).ToNot(HaveOccurred())
		Expect(code).To(Equal(detect.PassStatusCode))
		Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
			Provides: []buildplan.Provided{{Name: runtime.DotnetRuntime}},
		}))
	})

	it("passes when there is no valid runtimeconfig.json meaning that app is source based", func() {
		code, err := runDetect(factory.Detect)
		Expect(err).ToNot(HaveOccurred())
		Expect(code).To(Equal(detect.PassStatusCode))
		Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
			Provides: []buildplan.Provided{{Name: runtime.DotnetRuntime}},
		}))
	})

	it("passes when there is a valid runtimeconfig.json with comments", func() {
		runtimeConfigJSONPath := filepath.Join(factory.Detect.Application.Root, "appName.runtimeconfig.json")
		Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
{
  "runtimeOptions": {
    /*
    Multi line
    Comment
    */
    "tfm": "netcoreapp2.2",
    "framework": {
      "name": "Microsoft.NETCore.App",
      "version": "2.2.5"
    },
    // comment here ok?
    "configProperties": {
      "System.GC.Server": true
    }
  }
}
`), os.ModePerm)).To(Succeed())
		code, err := runDetect(factory.Detect)
		Expect(err).ToNot(HaveOccurred())
		Expect(code).To(Equal(detect.PassStatusCode))
		Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
			Provides: []buildplan.Provided{{Name: runtime.DotnetRuntime}},
			Requires: []buildplan.Required{{
				Name: runtime.DotnetRuntime,
				Version: "2.2.5",
				Metadata: buildplan.Metadata{"launch": true},
			}},
		}))
	})
}
