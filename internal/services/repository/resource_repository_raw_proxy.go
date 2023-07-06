package repository

import (
	nexus "github.com/datadrivers/go-nexus-client/nexus3"
	"github.com/datadrivers/go-nexus-client/nexus3/schema/repository"
	"github.com/datadrivers/terraform-provider-nexus/internal/schema/common"
	repositorySchema "github.com/datadrivers/terraform-provider-nexus/internal/schema/repository"
	"github.com/datadrivers/terraform-provider-nexus/internal/tools"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceRepositoryRawProxy() *schema.Resource {
	return &schema.Resource{
		Description: "Use this resource to create a raw proxy repository.",

		Create: resourceRawProxyRepositoryCreate,
		Delete: resourceRawProxyRepositoryDelete,
		Exists: resourceRawProxyRepositoryExists,
		Read:   resourceRawProxyRepositoryRead,
		Update: resourceRawProxyRepositoryUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			// Common schemas
			"id":     common.ResourceID,
			"name":   repositorySchema.ResourceName,
			"online": repositorySchema.ResourceOnline,
			// Proxy schemas
			"cleanup":                repositorySchema.ResourceCleanup,
			"http_client":            repositorySchema.ResourceHTTPClientWithPreemptiveAuth,
			"negative_cache_enabled": repositorySchema.ResourceNegativeCacheEnabled,
			"negative_cache_ttl":     repositorySchema.ResourceNegativeCacheTTL,
			"proxy":                  repositorySchema.ResourceProxy,
			"routing_rule":           repositorySchema.ResourceRoutingRule,
			"storage":                repositorySchema.ResourceStorage,
		},
	}
}

func getRawProxyRepositoryFromResourceData(resourceData *schema.ResourceData) repository.RawProxyRepository {
	httpClientConfig := resourceData.Get("http_client").([]interface{})[0].(map[string]interface{})
	/** negative_cache is an option an attribute of TypeList, which can not set default value for its since it's a limitation
	of terraform-plugin-sdk ref: https://github.com/hashicorp/terraform-plugin-sdk/issues/142
	It requires resource block to be existed for default value if its Elem to have effect, eg :
	negative_cache {}
	Its absence will cause the provider to crash
	**/
	negativeCacheEnabled := resourceData.Get("negative_cache_enabled").(bool)
	negativeCacheTTL := resourceData.Get("negative_cache_ttl").(int)
	negativeCacheConfig := map[string]interface{}{
		"enabled": negativeCacheEnabled,
		"ttl":     negativeCacheTTL,
	}
	proxyConfig := resourceData.Get("proxy").([]interface{})[0].(map[string]interface{})
	storageConfig := resourceData.Get("storage").([]interface{})[0].(map[string]interface{})

	repo := repository.RawProxyRepository{
		Name:   resourceData.Get("name").(string),
		Online: resourceData.Get("online").(bool),
		Storage: repository.Storage{
			BlobStoreName:               storageConfig["blob_store_name"].(string),
			StrictContentTypeValidation: storageConfig["strict_content_type_validation"].(bool),
		},
		HTTPClient: repository.HTTPClient{
			AutoBlock: httpClientConfig["auto_block"].(bool),
			Blocked:   httpClientConfig["blocked"].(bool),
		},
		NegativeCache: repository.NegativeCache{
			Enabled: negativeCacheConfig["enabled"].(bool),
			TTL:     negativeCacheConfig["ttl"].(int),
		},
		Proxy: repository.Proxy{
			ContentMaxAge:  proxyConfig["content_max_age"].(int),
			MetadataMaxAge: proxyConfig["metadata_max_age"].(int),
			RemoteURL:      proxyConfig["remote_url"].(string),
		},
	}

	if routingRule, ok := resourceData.GetOk("routing_rule"); ok {
		repo.RoutingRule = tools.GetStringPointer(routingRule.(string))
		repo.RoutingRuleName = tools.GetStringPointer(routingRule.(string))
	}

	cleanupList := resourceData.Get("cleanup").([]interface{})
	if len(cleanupList) > 0 && cleanupList[0] != nil {
		cleanupConfig := cleanupList[0].(map[string]interface{})
		if len(cleanupConfig) > 0 {
			policy_names, ok := cleanupConfig["policy_names"]
			if ok {
				repo.Cleanup = &repository.Cleanup{
					PolicyNames: tools.InterfaceSliceToStringSlice(policy_names.(*schema.Set).List()),
				}
			}
		}
	}

	if v, ok := httpClientConfig["authentication"]; ok {
		authList := v.([]interface{})
		if len(authList) == 1 && authList[0] != nil {
			authConfig := authList[0].(map[string]interface{})

			repo.HTTPClient.Authentication = &repository.HTTPClientAuthentication{
				NTLMDomain: authConfig["ntlm_domain"].(string),
				NTLMHost:   authConfig["ntlm_host"].(string),
				Type:       repository.HTTPClientAuthenticationType(authConfig["type"].(string)),
				Username:   authConfig["username"].(string),
				Password:   authConfig["password"].(string),
			}
		}
	}

	if v, ok := httpClientConfig["connection"]; ok {
		repo.HTTPClient.Connection = getHTTPClientConnection(v.([]interface{}))
	}

	return repo
}

func setRawProxyRepositoryToResourceData(repo *repository.RawProxyRepository, resourceData *schema.ResourceData) error {
	resourceData.SetId(repo.Name)
	resourceData.Set("name", repo.Name)
	resourceData.Set("online", repo.Online)

	if repo.RoutingRuleName != nil {
		resourceData.Set("routing_rule", repo.RoutingRuleName)
	} else if repo.RoutingRule != nil {
		resourceData.Set("routing_rule", repo.RoutingRule)
	}

	if err := resourceData.Set("storage", flattenStorage(&repo.Storage)); err != nil {
		return err
	}

	if err := resourceData.Set("http_client", flattenHTTPClient(&repo.HTTPClient, resourceData)); err != nil {
		return err
	}
	if err := resourceData.Set("negative_cache_enabled", repo.NegativeCache.Enabled); err != nil {
		return err
	}

	if err := resourceData.Set("negative_cache_ttl", repo.NegativeCache.TTL); err != nil {
		return err
	}

	if err := resourceData.Set("proxy", flattenProxy(&repo.Proxy)); err != nil {
		return err
	}

	if repo.Cleanup != nil {
		if err := resourceData.Set("cleanup", flattenCleanup(repo.Cleanup)); err != nil {
			return err
		}
	}
	return nil
}

func resourceRawProxyRepositoryCreate(resourceData *schema.ResourceData, m interface{}) error {
	client := m.(*nexus.NexusClient)

	repo := getRawProxyRepositoryFromResourceData(resourceData)

	if err := client.Repository.Raw.Proxy.Create(repo); err != nil {
		return err
	}
	resourceData.SetId(repo.Name)

	return resourceRawProxyRepositoryRead(resourceData, m)
}

func resourceRawProxyRepositoryRead(resourceData *schema.ResourceData, m interface{}) error {
	client := m.(*nexus.NexusClient)

	repo, err := client.Repository.Raw.Proxy.Get(resourceData.Id())
	if err != nil {
		return err
	}

	if repo == nil {
		resourceData.SetId("")
		return nil
	}

	return setRawProxyRepositoryToResourceData(repo, resourceData)
}

func resourceRawProxyRepositoryUpdate(resourceData *schema.ResourceData, m interface{}) error {
	client := m.(*nexus.NexusClient)

	repoName := resourceData.Id()
	repo := getRawProxyRepositoryFromResourceData(resourceData)

	if err := client.Repository.Raw.Proxy.Update(repoName, repo); err != nil {
		return err
	}

	return resourceRawProxyRepositoryRead(resourceData, m)
}

func resourceRawProxyRepositoryDelete(resourceData *schema.ResourceData, m interface{}) error {
	client := m.(*nexus.NexusClient)
	return client.Repository.Raw.Proxy.Delete(resourceData.Id())
}

func resourceRawProxyRepositoryExists(resourceData *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*nexus.NexusClient)

	repo, err := client.Repository.Raw.Proxy.Get(resourceData.Id())
	return repo != nil, err
}
