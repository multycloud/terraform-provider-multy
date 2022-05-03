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
		testString := strings.Replace(string(contents), "file(\"./ssh_key.pub\")", fmt.Sprintf("\"%s\"", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDFRQk+HkW4QXy1EdEd6BcCQcaT8pb/ySF98GvbXFTP/qZEnzl074SaBzefMP0zZi3N5vQD6tBWe/uxpZUKsHqkti+l6S3eR8Ols0E7jSpbLvfV+cBeNle7bdzH76V0SjUc3xEkAZNLrcKTNQgnot69ChE/Z5URwL1dMeD8GXATVtSH/AvGat3PSexkL75rWbCBXXmr+5/Re8kLSqYPf6WsLUbI6rIp3Okd1Kmo8pIHq9fqm/B9HSJjOXl08G2RC2H02+HIzRc6AIIqFBbPTQwjw5VEHaZiUC7tl5S117CpAx8oiv8njjR6+sNfEocjaPYl9ks/cVmpY1jCtEiP/5rBmfTSaBVm1BqAqbyLt+H2j7E/IzJBT1SWSy/tlk7r/E32b+JXCLfytNkoOlX7v3PrY9gy8927+4n0rmkLAHcglpXt93/Qqy81fv/QMmhLsnxL6JFrlvjx1X5GIiHvid3AG3K9Pm925whxMNN3HOLHxQPHghtB2iCgiv0DpU9sLjs= joao@Joaos-MBP"), -1)
		if err != nil {
			t.Errorf("unable to open %s, %s", fileName, err)
			continue
		}
		t.Run(fmt.Sprintf("%s_%s", filepath.Base(filepath.Dir(path)), os.Getenv("TF_VAR_cloud")), func(t *testing.T) {
			isError := strings.HasSuffix(filepath.Base(filepath.Dir(path)), "_failed")
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"multy": func() (tfprotov6.ProviderServer, error) {
						return tfsdk.NewProtocol6Server(&Provider{}), nil
					},
				},
				Steps: []resource.TestStep{getStep(testString, isError)},
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
