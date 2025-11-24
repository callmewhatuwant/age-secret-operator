# age-secret-operator

* Controller to decrypt secrets from a crd.
* Secrets must be encrypted via age.

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

## Helm

**Checkout my docs wich supports different chart versions**

<a href="https://age-secrets.com" target="_blank">age-secrets.com</a>

## kubectl apply -f

 * deploy rendered helm content
 * download and kustomize if u want.
 * service monitor is enabled crd is needed for that

```sh
https://raw.githubusercontent.com/callmewhatuwant/age-secret-operator/main/deploy/manifests/deploy.yaml
```
## To Deploy on the cluster with makefile

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
## Contributing

Contributions are welcome!  
You can contribute in several ways:

- **Report issues** (bugs, unexpected behavior, documentation gaps)
- **Request new features**
- **Submit pull requests**
- **Improve documentation or examples**

---

### Reporting Issues & Requesting Features

If you encounter a bug or want to request a new feature:

1. Open a new **GitHub Issue**
2. Describe the problem or request clearly
3. Add steps to reproduce (if applicable)
4. Include environment details (Go version, Kubernetes version, OS, etc.)

---

### Contributing Code (Pull Requests)

#### 1. Fork the repository
Create a fork of this project and work in a dedicated feature branch.

#### 2. Set up your environment
Ensure you have the required tools installed (Go, Kubernetes CLI tools, controller-runtime).

Run the following command to see all available development targets:

```sh
make help
``` 

**NOTE:** Run `make help` for more information on all potential `make` targets

Once you have implemented your changes and you believe the feature or fix is useful,  
**you are welcome to open a Pull Request**. 
Please include a short explanation of the change and why it improves the project.

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0.
See the LICENSE file for the full license text.


