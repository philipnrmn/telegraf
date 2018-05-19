# DC/OS Workloads Plugin

The DC/OS workloads plugin gathers metrics from applications running on top of statsd. When a container is started,
the mesos agent signals to the plugin that it should open a statsd server on an ephemeral port. The port number is 
passed back to the agent, which makes it available to the container's workload as an environment variable.

The workload plugin will tag all metrics which it receives on this port with the appropriate container ID. Downstream,
the [dcos_metadata](../../processors/dcos_metadata) plugin will add further relevant tags to these metrics. 

When the workload halts or is killed, the mesos agent signals to the plugin to close the statsd server. 

