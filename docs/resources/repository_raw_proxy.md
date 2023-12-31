---
page_title: "Resource nexus_repository_raw_proxy"
subcategory: "Repository"
description: |-
  Use this resource to create a raw proxy repository.
---
# Resource nexus_repository_raw_proxy
Use this resource to create a raw proxy repository.
## Example Usage
```terraform
resource "nexus_repository_raw_proxy" "raw_org" {
  name   = "raw-org"
  online = true

  storage {
    blob_store_name                = "default"
    strict_content_type_validation = true
  }

  proxy {
    remote_url       = "https://repo1.raw.org/raw2/"
    content_max_age  = 1440
    metadata_max_age = 1440
  }

  negative_cache_enabled = true
  negative_cache_ttl     = 1440


  http_client {
    blocked    = false
    auto_block = true
  }
}
```
<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `http_client` (Block List, Min: 1, Max: 1) HTTP Client configuration for proxy repositories (see [below for nested schema](#nestedblock--http_client))
- `name` (String) A unique identifier for this repository
- `proxy` (Block List, Min: 1, Max: 1) Configuration for the proxy repository (see [below for nested schema](#nestedblock--proxy))
- `storage` (Block List, Min: 1, Max: 1) The storage configuration of the repository (see [below for nested schema](#nestedblock--storage))

### Optional

- `cleanup` (Block List) Cleanup policies (see [below for nested schema](#nestedblock--cleanup))
- `negative_cache_enabled` (Boolean) Configuration of the negative cache handling, defaults to `false` if unset
- `negative_cache_ttl` (Number) Configuration of the negative cache handling, defaults is `1440` if unset
- `online` (Boolean) Whether this repository accepts incoming requests, defaults to `true` if unset
- `routing_rule` (String) The name of the routing rule assigned to this repository

### Read-Only

- `id` (String) Used to identify resource at nexus

<a id="nestedblock--http_client"></a>
### Nested Schema for `http_client`

Required:

- `auto_block` (Boolean) Whether to auto-block outbound connections if remote peer is detected as unreachable/unresponsive
- `blocked` (Boolean) Whether to block outbound connections on the repository

Optional:

- `authentication` (Block List, Max: 1) Authentication configuration of the HTTP client (see [below for nested schema](#nestedblock--http_client--authentication))
- `connection` (Block List, Max: 1) Connection configuration of the HTTP client (see [below for nested schema](#nestedblock--http_client--connection))

<a id="nestedblock--http_client--authentication"></a>
### Nested Schema for `http_client.authentication`

Required:

- `type` (String) Authentication type. Possible values: `ntlm` or `username`

Optional:

- `ntlm_domain` (String) The ntlm domain to connect
- `ntlm_host` (String) The ntlm host to connect
- `password` (String, Sensitive) The password used by the proxy repository
- `preemptive` (Boolean) Whether to use pre-emptive authentication. Use with caution, defaults to `false` if unset
- `username` (String) The username used by the proxy repository


<a id="nestedblock--http_client--connection"></a>
### Nested Schema for `http_client.connection`

Optional:

- `enable_circular_redirects` (Boolean) Whether to enable redirects to the same location (may be required by some servers), defaults to `false` if unset
- `enable_cookies` (Boolean) Whether to allow cookies to be stored and used, defaults to `false` if unset
- `retries` (Number) Total retries if the initial connection attempt suffers a timeout, defaults to `0` if unset
- `timeout` (Number) Seconds to wait for activity before stopping and retrying the connection
- `use_trust_store` (Boolean) Use certificates stored in the Nexus Repository Manager truststore to connect to external systems, defaults to `false` if unset
- `user_agent_suffix` (String) Custom fragment to append to User-Agent header in HTTP requests, defaults to `false` if unset



<a id="nestedblock--proxy"></a>
### Nested Schema for `proxy`

Required:

- `remote_url` (String) Location of the remote repository being proxied

Optional:

- `content_max_age` (Number) How long (in minutes) to cache artifacts before rechecking the remote repository, defaults to `1440` if unset
- `metadata_max_age` (Number) How long (in minutes) to cache metadata before rechecking the remote repository, defaults to `1440` if unset


<a id="nestedblock--storage"></a>
### Nested Schema for `storage`

Required:

- `blob_store_name` (String) Blob store used to store repository contents

Optional:

- `strict_content_type_validation` (Boolean) Whether to validate uploaded content's MIME type appropriate for the repository format, defaults to `true` if unset


<a id="nestedblock--cleanup"></a>
### Nested Schema for `cleanup`

Optional:

- `policy_names` (Set of String) List of policy names
## Import
Import is supported using the following syntax:
```shell
# import using the name of repository
terraform import nexus_repository_raw_proxy.raw_org raw-org
```
