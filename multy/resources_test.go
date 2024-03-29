package multy

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"golang.org/x/exp/slices"
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

	testNumber := 0

	for fileName, path := range allTests {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			t.Errorf("unable to open %s, %s", fileName, err)
			continue
		}
		testNumber += 1
		t.Run(fmt.Sprintf("%s_%s", filepath.Base(filepath.Dir(path)), os.Getenv("TF_VAR_cloud")), getTestFunc(path, string(contents), testNumber))
	}
}

var gcpTests = []string{
	"TestAccResources/network_interface_gcp",
	"TestAccResources/network_interface_security_group_association_gcp",
	"TestAccResources/public_ip_gcp",
}

func getTestFunc(path string, testString string, testNumber int) func(t *testing.T) {
	return func(t *testing.T) {
		if os.Getenv("TF_VAR_cloud") == "gcp" {
			if slices.Contains(gcpTests, t.Name()) {
				t.Skip("GCP not implemented yet for this resource")
			}
		}
		isError := strings.HasSuffix(filepath.Base(filepath.Dir(path)), "_failed")
		resource.ParallelTest(t, resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				"multy": providerserver.NewProtocol6WithError(&Provider{}),
			},
			Steps: []resource.TestStep{getStep(testString, isError, testNumber)},
		})

		t.Cleanup(func() {
			err := os.RemoveAll(filepath.Join(os.TempDir(), "multy", fmt.Sprintf("%s-%d", os.Getenv("USER_SECRET_PREFIX"), testNumber)))
			if err != nil {
				t.Logf("unable to cleanup: %s", err)
			}
		})
	}
}

func getProviderBlock(n int) string {
	secretPrefix, exists := os.LookupEnv("USER_SECRET_PREFIX")
	if !exists {
		panic("env var USER_SECRET_PREFIX is not set")
	}
	return fmt.Sprintf(`
provider "multy" {
  aws             = {}
  azure           = {}
  gcp             = { project = "multy-project" }
  server_endpoint = "localhost:8000"
  api_key = "%s-%d"
}
`, secretPrefix, n)
}

func getStep(config string, isError bool, n int) resource.TestStep {
	step := resource.TestStep{
		Config: config + getProviderBlock(n),
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
