# storage-pool-migrator

An open-source k8s operator to migrate workloads to Gen4 GCP instances. 

## Overview

On November 2025, Google Cloud introduced Gen4 instances, which offer a significant performance boost over Gen2/Gen3 instances. However, the following challenges arise when migrating workloads to Gen4 instances:

1. For Gen2/Gen3 instances, the supported storage classes are limited to `pd-standard`, `pd-balanced` and `pd-ssd`. But the Gen4 instances do not support these storage classes.
2. The supported storage classes for Gen4 instances are `hyperdisk-balanced`, `hyperdisk-throughput`, `hyperdisk-extreme`, `hyperdisk-ml`and `hyperdisk-balanced-ha`.
3. So, the migration process requires manual intervention to create new storage classes; update the node selectors, tolerations for deployments/statefulsets to use the new Gen4 nodes; Migrate data from old PVCs to new PVCs using backup/restore tools.


## Operator Workflow

Following are the steps `storage-pool-migrator` performs to migrate a workload to Gen4 instances:

1. The operator watches for deployments with annotations `storagepool.codeamenity.in/enabled: "true"`.
2. Scale down the deployment to 0 replicas.
3. Create new storage classes for Gen4 instances if it does not exist.
4. Update the node selectors and tolerations for the deployment to use Gen4 nodes.
5. Create volume snapshots for the PVCs associated with the deployment. Ignore if the PVCs are already migrated to new storage classes.
6. Delete/Rename the old PVCs.
7. Create new PVCs using the new storage classes for Gen4 instances.
8. Restore the data from the volume snapshots to the new PVCs.
9. Scale up the deployment to the original number of replicas.