package repository_test

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
	"text/template"

	"github.com/datadrivers/terraform-provider-nexus/internal/acceptance"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nduyphuong/go-nexus-client/nexus3/schema/repository"
)

func testAccResourceRepositoryBowerHosted() repository.BowerHostedRepository {
	writePolicy := repository.StorageWritePolicyAllow

	return repository.BowerHostedRepository{
		Name:   fmt.Sprintf("test-repo-%s", acctest.RandString(10)),
		Online: true,
		Storage: repository.HostedStorage{
			BlobStoreName:               "default",
			StrictContentTypeValidation: true,
			WritePolicy:                 &writePolicy,
		},
		Cleanup: &repository.Cleanup{
			PolicyNames: []string{"cleanup-weekly"},
		},
		Component: &repository.Component{
			ProprietaryComponents: true,
		},
	}
}

func testAccResourceRepositoryBowerHostedConfig(repo repository.BowerHostedRepository) string {
	buf := &bytes.Buffer{}
	resourceRepositoryBowerHostedTemplate := template.Must(template.New("BowerHostedRepository").Funcs(acceptance.TemplateFuncMap).Parse(acceptance.TemplateStringRepositoryBowerHosted))
	if err := resourceRepositoryBowerHostedTemplate.Execute(buf, repo); err != nil {
		panic(err)
	}
	return buf.String()
}

func TestAccResourceRepositoryBowerHosted(t *testing.T) {
	repo := testAccResourceRepositoryBowerHosted()
	resourceName := "nexus_repository_bower_hosted.acceptance"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acceptance.AccPreCheck(t) },
		Providers: acceptance.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRepositoryBowerHostedConfig(repo),
				Check: resource.ComposeTestCheckFunc(
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "id", repo.Name),
						resource.TestCheckResourceAttr(resourceName, "name", repo.Name),
						resource.TestCheckResourceAttr(resourceName, "online", strconv.FormatBool(repo.Online)),
					),
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "storage.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "storage.0.blob_store_name", repo.Storage.BlobStoreName),
						resource.TestCheckResourceAttr(resourceName, "storage.0.strict_content_type_validation", strconv.FormatBool(repo.Storage.StrictContentTypeValidation)),
						resource.TestCheckResourceAttr(resourceName, "storage.0.write_policy", string(*repo.Storage.WritePolicy)),
						resource.TestCheckResourceAttr(resourceName, "cleanup.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "cleanup.0.policy_names.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "cleanup.0.policy_names.0", repo.Cleanup.PolicyNames[0]),
						resource.TestCheckResourceAttr(resourceName, "component.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "component.0.proprietary_components", strconv.FormatBool(repo.Component.ProprietaryComponents)),
					),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     repo.Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
