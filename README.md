# Temporarily scale down all deployments in a namespace

### What it does

Downscale all deployments in a namespace to 0 replicas and back to their
previous state. Useful when you need to tear everything down and bring
it back in a namespace.

### Usage

Downscale all deployments in a namespace to 0 replicas:

```bash
szero down -n <namespace>
```

Upscale all deployments in a namespace to their previous state:

```bash
szero up -n <namespace>
```

Restart all deployments in a namespace

```bash
szero restart -n <namespace>
```
