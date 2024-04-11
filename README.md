# Temporarily scale down all deployments in a namespace

### What it does

Downscale all deployments in a namespace to 0 replicas and back to their
previous state. Useful when you need to tear everything down and bring
it back in a namespace.

### Usage

#### Downscale all deployments in a namespace to 0 replicas:

```bash
szero down -n <namespace> -n <another_namespace>
```

#### Upscale all deployments in a namespace to their previous state:

```bash
szero up -n <namespace> -n <another_namespace>
```

#### Restart all deployments in a namespace

```bash
szero restart -n <namespace> -n <another_namespace>
```

#### Use a different kubeconfig file

```bash
szero down -n <namespace> --kubeconfig <path_to_kubeconfig>
```

#### Use a different context

```bash
szero down -n <namespace> --context <context_name>
```

## Completions
Command line completions are available under the `completions` subcommand.
For example, to enable bash completions, run:
```bash
source <(szero completion bash)
```
