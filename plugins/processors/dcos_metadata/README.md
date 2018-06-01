# DC/OS Metadata Plugin

The DC/OS metdata processor plugin tracks state on the mesos agent, and decorates every metric which passes through it
from the [dcos_container](../../input/dcos_container) and [dcos_workload](../../input/dcos_workload) with appropriate
metadata in the form of DC/OS primitives. 

### Configuration

```toml
[[processors.dcos_metadata]]
```

### Tags:

This process adds the following tags to any metric with a container_id tag set:

 - `task_name` - the name of the task associated with this container

### Fields:

This processor does not add fields.
