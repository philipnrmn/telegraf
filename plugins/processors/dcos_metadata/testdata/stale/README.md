# Scenario: Normal

- Given that a task is _no longer_ running on the cluster
- And that task's information is cached
- When container metrics are retrieved
- Then that task's container metrics should not be present
