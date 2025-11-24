# age-secret-operator

* Controller to decrypt secrets from a crd.
* Secrets must be encrypted via age.

**checkout helm docs: <a href="https://age-secrets.com" target="_blank">age-secrets.com</a>**

## Description

The Controller can be installed via helm or manifests (helm is prefered).
Also the Crd must be installed.
After the deplyoment, a job runs wich creats a secret, with an age private and public key.
This secret should be used to encrypt your secret values.
Every first of the month a new secret for encrypting will be generated.
The controller can use all keys in his namespace to decrypt the crd component in every namespace.
Please not if you delete a secret you will not be able to decrypt the crd resource wich got encrypted
with these keys. 

## Getting Started

### Prerequisites
- go version v1.24.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster with makefile

* If you wanna build the image by yourself.
* I recommend to use my prebuilt images and to use the helm section

**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build-(alpine,op,gui) docker-push IMG=<some-registry>/age-secret-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=dockerhub.io/callmewhatuwant/age-secret-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Helm

**Checkout my docs wich supports different chart versions**

<a href="https://age-secrets.com" target="_blank">age-secrets.com</a>


## Project Distribution: Todo
// I have to change this part </br>
Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/age-secret-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/age-secret-operator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0.
See the LICENSE file for the full license text.


