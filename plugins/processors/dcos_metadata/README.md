# DC/OS Metadata Plugin

The DC/OS metdata processor plugin tracks state on the mesos agent, and decorates every metric which passes through it
from the [dcos_container](../../input/dcos_container) and [dcos_workload](../../input/dcos_workload) with appropriate
metadata in the form of DC/OS primitives. 

