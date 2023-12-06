# Temporarily scale down all deployments in a namespace

### Usage

Use the following command to scale down all deployments in a namespace to 0 replicas.
Default namespace is `default`.
```bash
szero -n namespace
```
The application will prompt for confirmation before scaling back up the deployments again.
