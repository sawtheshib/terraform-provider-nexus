package repository

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nexus "github.com/nduyphuong/go-nexus-client/nexus3"
	"github.com/nduyphuong/go-nexus-client/nexus3/schema/repository"
	"github.com/nduyphuong/terraform-provider-nexus/internal/schema/common"
	repositorySchema "github.com/nduyphuong/terraform-provider-nexus/internal/schema/repository"
)

func ResourceRepositoryRGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Use this resource to create a group r repository.",

		Create: resourceRGroupRepositoryCreate,
		Delete: resourceRGroupRepositoryDelete,
		Exists: resourceRGroupRepositoryExists,
		Read:   resourceRGroupRepositoryRead,
		Update: resourceRGroupRepositoryUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			// Common schemas
			"id":     common.ResourceID,
			"name":   repositorySchema.ResourceName,
			"online": repositorySchema.ResourceOnline,
			// Group schemas
			"group":   repositorySchema.ResourceGroup,
			"storage": repositorySchema.ResourceStorage,
		},
	}
}

func getRGroupRepositoryFromResourceData(resourceData *schema.ResourceData) repository.RGroupRepository {
	storageConfig := resourceData.Get("storage").([]interface{})[0].(map[string]interface{})
	groupConfig := resourceData.Get("group").([]interface{})[0].(map[string]interface{})
	groupMemberNamesInterface := groupConfig["member_names"].([]interface{})
	groupMemberNames := make([]string, 0)
	for _, v := range groupMemberNamesInterface {
		groupMemberNames = append(groupMemberNames, v.(string))
	}

	repo := repository.RGroupRepository{
		Name:   resourceData.Get("name").(string),
		Online: resourceData.Get("online").(bool),
		Storage: repository.Storage{
			BlobStoreName:               storageConfig["blob_store_name"].(string),
			StrictContentTypeValidation: storageConfig["strict_content_type_validation"].(bool),
		},
		Group: repository.Group{
			MemberNames: groupMemberNames,
		},
	}

	return repo
}

func setRGroupRepositoryToResourceData(repo *repository.RGroupRepository, resourceData *schema.ResourceData) error {
	resourceData.SetId(repo.Name)
	resourceData.Set("name", repo.Name)
	resourceData.Set("online", repo.Online)

	if err := resourceData.Set("storage", flattenStorage(&repo.Storage)); err != nil {
		return err
	}

	if err := resourceData.Set("group", flattenGroup(&repo.Group)); err != nil {
		return err
	}

	return nil
}

func resourceRGroupRepositoryCreate(resourceData *schema.ResourceData, m interface{}) error {
	client := m.(*nexus.NexusClient)

	repo := getRGroupRepositoryFromResourceData(resourceData)

	if err := client.Repository.R.Group.Create(repo); err != nil {
		return err
	}
	resourceData.SetId(repo.Name)

	return resourceRGroupRepositoryRead(resourceData, m)
}

func resourceRGroupRepositoryRead(resourceData *schema.ResourceData, m interface{}) error {
	client := m.(*nexus.NexusClient)

	repo, err := client.Repository.R.Group.Get(resourceData.Id())
	if err != nil {
		return err
	}

	if repo == nil {
		resourceData.SetId("")
		return nil
	}

	return setRGroupRepositoryToResourceData(repo, resourceData)
}

func resourceRGroupRepositoryUpdate(resourceData *schema.ResourceData, m interface{}) error {
	client := m.(*nexus.NexusClient)

	repoName := resourceData.Id()
	repo := getRGroupRepositoryFromResourceData(resourceData)
	repo1, err := client.Repository.R.Group.Get(resourceData.Id())
	if err != nil {
		return err
	}
	if reflect.DeepEqual(repo1.Group.MemberNames, repo.Group.MemberNames) {
		return nil
	}
	if err := client.Repository.R.Group.Update(repoName, repo); err != nil {
		return err
	}

	return resourceRGroupRepositoryRead(resourceData, m)
}

func resourceRGroupRepositoryDelete(resourceData *schema.ResourceData, m interface{}) error {
	client := m.(*nexus.NexusClient)
	return client.Repository.R.Group.Delete(resourceData.Id())
}

func resourceRGroupRepositoryExists(resourceData *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*nexus.NexusClient)

	repo, err := client.Repository.R.Group.Get(resourceData.Id())
	return repo != nil, err
}
