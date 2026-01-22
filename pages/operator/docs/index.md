# Docs

Welcome to the documentation.</br>
The code and chart is avalible at [github.com](https://github.com/callmewhatuwant/age-secret-operator){target="_blank"}.

## Getting started

* create crd ressource
* crd has to be applied before using helm to install

```bash 
kubectl apply -f https://raw.githubusercontent.com/callmewhatuwant/age-secret-operator/refs/heads/main/config/crd/bases/security.age.io_agesecrets.yaml
```

* install

```bash
helm repo add age-secrets-operator \
https://age-secrets.com
helm install age-secrets-operator age-secrets-operator/age-secrets \
--namespace age-system --create-namespace
```

* check install

```bash
kubectl wait --for=condition=Ready pods --all -n age-system
```

* uninstall

```bash
helm uninstall -n age-system age-secrets-operator
kubectl delete namespace age-system
```

## First secret

* install age

```bash
sudo apt install age
```

* get key

```bash
LATEST=$(kubectl get secrets -n age-system --no-headers -o custom-columns=":metadata.name" \
  | grep '^age-key-' | sort | tail -n1)

PUBLIC_KEY=$(kubectl get secret "$LATEST" -n age-system -o jsonpath='{.data.public}' | base64 --decode)

## public key:
echo "$PUBLIC_KEY"
```

* encrypt with ur public key

```bash
echo test123 > secret.txt

age --armor -r "$PUBLIC_KEY" secret.txt
```

* create namespace for testing the crd

```bash
kubectl create ns test
```

* exmaple secret crd ressource
* change password: value with your value from the age command

```yaml
apiVersion: security.age.io/v1alpha1
kind: AgeSecret
metadata:
  name: db-passwd
  namespace: test
spec:
  encryptedData:
    password: |
      -----BEGIN AGE ENCRYPTED FILE-----
      YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSBWbHhqcGhyZ0ZSbXhQZXJ1
      aU1kL1NmZjYyaU9JQXlQazBuekdmMk8ySkYwCkloMGJxR0lXVG0yM2FXV3hrT3BI
      OXVwdzhrYWtGU0hwTUtLTHN5dzJBTGsKLS0tIEc0V1JmTUVpWkZuNGFGWXJJV3ow
      cWZpL09JTnFCVFFZbXRFQUY2QTdTbm8KdkZOvCXRqENpCw9ncrVP+qzDBTKwntfi
      ihgfMGuoy3Q37Dkqsw==
      -----END AGE ENCRYPTED FILE-----
  # recipients:
  # define public key used as value if u want
  #   - string
```

* verify

```bash
kubectl get secret -n test
```

## Helm Options

```yaml
## name override
fullnameOverride: age-secret-controller
  
## namespaces in wich new keys will be generated
## controller will also check in them for keys to decrypt  
keyNamespaces: "age-secrets"

## controller values  
ageSecretController:

## leader election
  leaderElection:
    enabled: true
    namespace: age-system

  ## replicas for ha
  replicas: 3

  controller:
    ## image
    image:
      repository: callmewhatuwant/age-secrets-operator
      tag: 0.0.02
    imagePullPolicy: IfNotPresent

    ## resources
    resources:
      limits:
        cpu: 200m
        memory: 128Mi
      requests:
        cpu: 100m
        memory: 64Mi

    ## security
    containerSecurityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      seccompProfile:
        type: "RuntimeDefault"
      runAsNonRoot: true
      privileged: false
      runAsUser: 65532
      runAsGroup: 65532

## prometheus
metricsService:
  bindAddress: 8443
  secure: true
  auth: false
  
  type: ClusterIP
  ports:
    - port: 8443
      name: metrics
      targetPort: 8443

## monitor for prometheus
ServiceMonitor:
  enabled: false
  endpoints:
    - port: metrics
      scheme: https
      interval: 30s
      path: /metrics
      tlsConfig:
        insecureSkipVerify: true
        serverName: localhost
        
## job
ageKeyRotation:
  schedule: "0 0 1 * *"
  ## initial key
  initialRun:
    enabled: true

  ## image for cron and init job
  image:
    repository: callmewhatuwant/age-job
    tag: "3.23.0"
    pullPolicy: IfNotPresent

## gui
ageGui:
  enabled: true
  replicas: 1

  # image for gui
  image:
    repository: callmewhatuwant/age-gui
    tag: "alpine3.22-perl"
    pullPolicy: IfNotPresent

  resources:
    limits:
      cpu: 200m
      memory: 128Mi
    requests:
      cpu: 100m
      memory: 64Mi
    
  # strategy for updating
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: "25%"
      maxUnavailable: "25%"

  #sec context
  containerSecurityContext:
    allowPrivilegeEscalation: false
    capabilities:
      drop:
        - ALL
    seccompProfile:
      type: "RuntimeDefault"
    runAsNonRoot: true
    privileged: false
    runAsUser: 101
    runAsGroup: 101

  # service for gui
  service:
    type: ClusterIP
    ports:
      - name: http
        port: 80
        targetPort: 8080
        protocol: TCP

  # ingress for gui
  ingress:
    enabled: false
    host: age-gui.local
    ingressClassName: nginx
    annotations: {}
    tls: []
      # - hosts:
      #     - age-gui.local
      #   secretName: age-gui-tl
```

## Enhancements

* Open a merge request if you want to contribute to the project.  
* The project just started, so there’s probably a lot to improve.  
* Please don’t be too harsh on me. 🙂

## Issue

* Found a bug or have a feature request?
* Please open an issue on [github.com](https://github.com/callmewhatuwant/age-secret-operator){target="_blank"}.

## Other projects by me

* my portfolio [portfolio-nick.de](https://www.portfolio-nick.de){target="_blank"}

## Support me if you want

BTC:

```bash
bc1q7zgprykqzj4vprzxzafy5lskhpv7qau9p7a28r
```

Solana:
```bash
B6aGswkR4tpYDCaLny4B1rZWwQNrDk4dEvpEGjJw3GGG
```

