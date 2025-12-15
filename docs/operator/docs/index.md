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

echo "$PUBLIC_KEY"

echo test123 > secret.txt
```

* encrypt with ur public key

```bash
age --armor -r "$PUBLIC_KEY" secret.txt
```

* exmaple secret crd ressource

```yaml
apiVersion: security.age.io/v1alpha1
kind: AgeSecret
metadata:
  name: db-passwd
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
```

* verify

```bash
kubectl get secret -n age-system
```

## Helm Options

```yaml
## name override
fullnameOverride: age-secret-controller
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
      tag: 0.0.5
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
      runAsNonRoot: true
      runAsUser: 65532

## prometheus
metricsService:
  type: ClusterIP
  ports:
    - port: 8080
      name: metrics
      targetPort: 8080

## monitor for prometheus
ServiceMonitor:
  enabled: true
  endpoints:
    - port: metrics
      interval: 30s
      path: /metrics

## job
ageKeyRotation:
  schedule: "0 0 1 * *"

  ## initial key
  initialRun:
    enabled: true

  ## image for cron and init job
  image:
    repository: callmewhatuwant/age-job
    tag: "3.22.2"
    pullPolicy: IfNotPresent

## gui
ageGui:
  enabled: false
  replicas: 1

  # image for gui
  image:
    repository: callmewhatuwant/age-gui
    tag: "alpine3.22-perl"
    pullPolicy: IfNotPresent
    
  # strategy for updating
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: "25%"
      maxUnavailable: "25%"

  # sec context
  containerSecurityContext:
    allowPrivilegeEscalation: false
    capabilities:
      drop:
        - ALL
    runAsNonRoot: true
    runAsUser: 101

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

