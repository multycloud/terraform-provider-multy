package multy

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestAccResources(t *testing.T) {
	allTests := map[string]string{}

	err := filepath.WalkDir("../tests/resources", func(path string, info os.DirEntry, err error) error {
		if info.IsDir() || filepath.Base(path) != "main.tf" || filepath.Ext(path) != ".tf" || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		t.Logf("found %s", path)
		ext := filepath.Ext(path)
		base := strings.TrimSuffix(path, ext)
		if _, ok := allTests[base]; !ok {
			allTests[base] = path
		}

		return nil
	})

	if err != nil {
		t.Fatalf("unable to get test files, %s", err)
	}

	for fileName, path := range allTests {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			t.Errorf("unable to open %s, %s", fileName, err)
			continue
		}
		t.Run(filepath.Base(filepath.Dir(path)), func(t *testing.T) {
			isError := strings.HasSuffix(filepath.Base(filepath.Dir(path)), "_failed")
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"multy": func() (tfprotov6.ProviderServer, error) {
						return tfsdk.NewProtocol6Server(&Provider{}), nil
					},
				},
				Steps: []resource.TestStep{getStep(string(contents), isError)},
			})
		})
	}

}

var providerBlock = `
provider "multy" {
  aws             = {}
  azure           = {}
  server_endpoint = "localhost:8000"
}
`

func getStep(config string, isError bool) resource.TestStep {
	step := resource.TestStep{
		Config: config + providerBlock,
	}

	if isError {
		step.ExpectError = regexp.MustCompile(".*")
	} else {
		step.Check = func(state *terraform.State) error {
			if !state.HasResources() {
				return fmt.Errorf("no resources")
			}
			return nil
		}
	}

	return step
}
