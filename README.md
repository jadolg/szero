# Temporarily scale down all deployments in a namespace

### What it does

This application iterates over all deployments in a namespace and scales them
down to 0 replicas. It then prompts for confirmation before scaling them back
up again to their original scale.


### Usage

Use the following command to scale down all deployments in a namespace to 0
replicas. Default namespace is `default`.
```bash
szero -n namespace
```
The application will prompt for confirmation before scaling back up the 
deployments again.
