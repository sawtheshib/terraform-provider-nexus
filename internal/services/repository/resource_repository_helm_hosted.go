package repository

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nexus "github.com/nduyphuong/go-nexus-client/nexus3"
	"github.com/nduyphuong/go-nexus-client/nexus3/schema/repository"
	"github.com/nduyphuong/terraform-provider-nexus/internal/schema/common"
	repositorySchema "github.com/nduyphuong/terraform-provider-nexus/internal/schema/repository"
	"github.com/nduyphuong/terraform-provider-nexus/internal/tools"
)

func ResourceRepositoryHelmHosted() *schema.Resource {
	return &schema.Resource{
		Description: "Use this resource to create a hosted helm repository.",

		Create: resourceHelmHostedRepositoryCreate,
		Delete: resourceHelmHostedRepositoryDelete,
		Exists: resourceHelmHostedRepositoryExists,
		Read:   resourceHelmHostedRepositoryRead,
		Update: resourceHelmHostedRepositoryUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			// Common schemas
			"id":     common.ResourceID,
			"name":   repositorySchema.ResourceName,
			"online": repositorySchema.ResourceOnline,
			// Hosted schemas
			"cleanup":   repositorySchema.ResourceCleanup,
			"component": repositorySchema.ResourceComponent,
			"storage":   repositorySchema.ResourceHostedStorage,
		},
	}
}

func getHelmHostedRepositoryFromResourceData(resourceData *schema.ResourceData) repository.HelmHostedRepository {
	storageConfig := resourceData.Get("storage").([]interface{})[0].(map[string]interface{})
	writePolicy := repository.StorageWritePolicy(storageConfig["write_policy"].(string))

	repo := repository.HelmHostedRepository{
		Name:   resourceData.Get("name").(string),
		Online: resourceData.Get("online").(bool),
		Storage: repository.HostedStorage{
			BlobStoreName:               storageConfig["blob_store_name"].(string),
			StrictContentTypeValidation: storageConfig["strict_content_type_validation"].(bool),
			WritePolicy:                 &writePolicy,
		},
	}

	cleanup, exists := resourceData.GetOk("cleanup")
	if exists {
		cleanupConfig := cleanup.([]interface{})[0].(map[string]interface{})
		if len(cleanupConfig) > 0 {
			policy_names, ok := cleanupConfig["policy_names"]
			if ok {
				repo.Cleanup = &repository.Cleanup{
					PolicyNames: tools.InterfaceSliceToStringSlice(policy_names.(*schema.Set).List()),
				}
			}
		}
	}

	componentList := resourceData.Get("component").([]interface{})
	if len(componentList) > 0 && componentList[0] != nil {
		componentConfig := componentList[0].(map[string]interface{})
		if len(componentConfig) > 0 {
			repo.Component = &repository.Component{
				ProprietaryComponents: componentConfig["proprietary_components"].(bool),
			}
		}
	}

	return repo
}

func setHelmHostedRepositoryToResourceData(repo *repository.HelmHostedRepository, resourceData *schema.ResourceData) error {
	resourceData.SetId(repo.Name)
	resourceData.Set("name", repo.Name)
	resourceData.Set("online", repo.Online)

	if err := resourceData.Set("storage", flattenHostedStorage(&repo.Storage)); err != nil {
		return err
	}

	if repo.Cleanup != nil {
		if err := resourceData.Set("cleanup", flattenCleanup(repo.Cleanup)); err != nil {
			return err
		}
	}

	if repo.Component != nil {
		if err := resourceData.Set("component", flattenComponent(repo.Component)); err != nil {
			return err
		}
	}

	return nil
}

func resourceHelmHostedRepositoryCreate(resourceData *schema.ResourceData, m interface{}) error {
	client := m.(*nexus.NexusClient)

	repo := getHelmHostedRepositoryFromResourceData(resourceData)

	if err := client.Repository.Helm.Hosted.Create(repo); err != nil {
		return err
	}
	resourceData.SetId(repo.Name)

	return resourceHelmHostedRepositoryRead(resourceData, m)
}

func resourceHelmHostedRepositoryRead(resourceData *schema.ResourceData, m interface{}) error {
	client := m.(*nexus.NexusClient)

	repo, err := client.Repository.Helm.Hosted.Get(resourceData.Id())
	if err != nil {
		return err
	}

	if repo == nil {
		resourceData.SetId("")
		return nil
	}

	return setHelmHostedRepositoryToResourceData(repo, resourceData)
}

func resourceHelmHostedRepositoryUpdate(resourceData *schema.ResourceData, m interface{}) error {
	client := m.(*nexus.NexusClient)

	repoName := resourceData.Id()
	repo := getHelmHostedRepositoryFromResourceData(resourceData)

	if err := client.Repository.Helm.Hosted.Update(repoName, repo); err != nil {
		return err
	}

	return resourceHelmHostedRepositoryRead(resourceData, m)
}

func resourceHelmHostedRepositoryDelete(resourceData *schema.ResourceData, m interface{}) error {
	client := m.(*nexus.NexusClient)
	return client.Repository.Helm.Hosted.Delete(resourceData.Id())
}

func resourceHelmHostedRepositoryExists(resourceData *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*nexus.NexusClient)

	repo, err := client.Repository.Helm.Hosted.Get(resourceData.Id())
	return repo != nil, err
}
