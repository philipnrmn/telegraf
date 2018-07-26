# DC/OS Containers Plugin

The DC/OS containers plugin gathers metrics about the resource consumption of
containers being run by the local Mesos agent. 

### Configuration:

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage dcos_containers`.

```toml
# Telegraf plugin for gathering resource metrics about mesos containers
[[inputs.dcos_containers]]
  ## The URL of the mesos agent
  mesos_agent_url = "http://localhost:5051"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

With minimal configuration, this plugin expects the cluster to be in permissive
mode. Strict mode requires TLS configuration. 

### Metrics:

<!-- TODO: consider including
 - processes
 - threads
 -->

 - cpus
   - fields:
     - user_time_secs
     - system_time_secs
     - limit
     - nr_periods
     - nr_throttled
     - throttled_time_secs

 - mem
   - fields:
     - total_bytes
     - total_memsw_bytes
     - limit_bytes
     - soft_limit_bytes
     - cache_bytes
     - rss_bytes
     - mapped_file_bytes
     - swap_bytes
     - unevictable_bytes
     - low_pressure_counter
     - medium_pressure_counter
     - critical_pressure_counter

 - disk
   - fields:
     - limit_bytes
     - used_bytes

 - net
   - tags:
     - rx_tx
   - fields:
     - packets
     - bytes
     - errors
     - dropped

 - blkio
   - tags:
     - policy <!-- cfq/cfq_recursive/throttling -->
     - device <!-- eg /dev/sda1 -->
   - fields:
     - serviced
     - service_bytes
     - service_time
     - merged
     - queued
     - wait_time
 
### Tags:

All metrics have the following tags:

 - container_id (a unique identifer given by mesos to the workload's container)
 - task_name
 - executor_name
 - service_name

### Example Output:

<!-- TODO: expand with all metrics -->
```
$ telegraf --config dcos.conf --input-filter dcos_container --test
* Plugin: dcos_container
    cpus,host=172.17.8.102,task_name=nginx-server-0,executor_name=nginx-server,service_name=nginx,container_id=12377985-615c-4a1a-a491-721ce7cd807a user_time_secs=10,system_time_secs=1,limit=4,nr_periods=11045,nr_throttled=132,throttled_time_seconds=1 1453831884664956455
```
