# kubernetes-metadata-injector

kubernetes-metadata-injector is a [Kubernetes admission webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) designed to enhance service objects within your Kubernetes clusters dynamically. Inspired by and based on [slackhq/simple-kubernetes-webhook](https://github.com/slackhq/simple-kubernetes-webhook), this project introduces a powerful way to manage annotations in Kubernetes services dynamically.

## Overview

The Kubernetes Metadata Injector operates as a Mutation Webhook, specifically targeting service creation and update operations. Its primary function is to inject annotations into services based on the values defined in a ConfigMap, which is referenced by a key within the service's annotations.

Utilizing a configurable `TRIGGER_ANNOTATION_PREFIX` (should be passed as env var), the webhook scans for annotations that match this prefix. These annotations should detail:
- The annotation to be injected, identified within the key as `<TRIGGER_ANNOTATION_PREFIX>.annotation.<injectable-annotation-key>`.
- The ConfigMap and ConfigMap key to fetch the annotation value from, specified as `<configmap-name>:<configmap-key>` within the value.

Upon detecting these specifications, the webhook injects a new annotation into the service, or updates an existing one, with the value retrieved from the specified ConfigMap key within the same namespace as the service.


## Example

Given a ConfigMap named `configmap` within the `default` namespace:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap
  namespace: default
data:
  key: configmap-value
```

And a service named `example-svc`, also in the `default` namespace, with an annotation specifying the injection:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: example-svc
  namespace: default
  annotations:
    moveax.injector.annotation.custom-annotation/subpath: "configmap:key"
spec:
  ports:
    - name: http
      protocol: TCP
      appProtocol: http
      port: 80
      targetPort: http
  selector:
    app.kubernetes.io/instance: example
    app.kubernetes.io/name: example
```

The resulting service after the webhook's injection will include the new annotation with the value from the ConfigMap:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: example-svc
  namespace: default
  annotations:
    moveax.injector.annotation.custom-annotation/subpath: "configmap:key"
    custom-annotation/subpath: configmap-value
spec:
  ports:
    - name: http
      protocol: TCP
      appProtocol: http
      port: 80
      targetPort: http
  selector:
    app.kubernetes.io/instance: example
    app.kubernetes.io/name: example
```

This feature allows for dynamic configuration and management of service annotations, streamlining operations and enabling more flexible deployment strategies within Kubernetes environments.

## Development
### Installation
This project can fully run locally and includes automation to deploy a local Kubernetes cluster (using Kind).

#### Requirements
* Docker
* kubectl
* [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
* Go >=1.16 (optional)

### Usage
#### Create Cluster
First, we need to create a Kubernetes cluster:
```
❯ make cluster

🔧 Creating Kubernetes cluster...
kind create cluster --config dev/manifests/kind/kind.cluster.yaml
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.21.1) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind

Have a nice day! 👋
```

Make sure that the Kubernetes node is ready:
```
❯ kubectl get nodes
NAME                 STATUS   ROLES                  AGE     VERSION
kind-control-plane   Ready    control-plane,master   3m25s   v1.21.1
```

And that system pods are running happily:
```
❯ kubectl -n kube-system get pods
NAME                                         READY   STATUS    RESTARTS   AGE
coredns-558bd4d5db-thwvj                     1/1     Running   0          3m39s
coredns-558bd4d5db-w85ks                     1/1     Running   0          3m39s
etcd-kind-control-plane                      1/1     Running   0          3m56s
kindnet-84slq                                1/1     Running   0          3m40s
kube-apiserver-kind-control-plane            1/1     Running   0          3m54s
kube-controller-manager-kind-control-plane   1/1     Running   0          3m56s
kube-proxy-4h6sj                             1/1     Running   0          3m40s
kube-scheduler-kind-control-plane            1/1     Running   0          3m54s
```

#### Deploy Admission Webhook
To configure the cluster to use the admission webhook and to deploy said webhook, simply run:
```
❯ make deploy

📦 Building kubernetes-metadata-injector Docker image...
docker build -t kubernetes-metadata-injector:latest .
[+] Building 14.3s (13/13) FINISHED
...

📦 Pushing admission-webhook image into Kind's Docker daemon...
kind load docker-image kubernetes-metadata-injector:latest
Image: "kubernetes-metadata-injector:latest" with ID "sha256:30d69761cf53433b517aa3c578e577a5e337f66553628865cd1b0c0df27c09cf" not yet present on node "kind-control-plane", loading...

⚙️  Applying cluster config...
kubectl apply -f dev/manifests/cluster-config/
namespace/apps created
mutatingwebhookconfiguration.admissionregistration.k8s.io/kubernetes-metadata-injector.acme.com created

🚀 Deploying kubernetes-metadata-injector...
kubectl apply -f dev/manifests/webhook/
serviceaccount/kubernetes-metadata-injector created
clusterrole.rbac.authorization.k8s.io/kubernetes-metadata-injector created
clusterrolebinding.rbac.authorization.k8s.io/kubernetes-metadata-injector created
deployment.apps/kubernetes-metadata-injector created
service/kubernetes-metadata-injector created
secret/kubernetes-metadata-injector-tls created
```

Then, make sure the admission webhook pod is running (in the `default` namespace):
```
❯ kubectl get pods
NAME                                        READY   STATUS    RESTARTS   AGE
kubernetes-metadata-injector-77444566b7-wzwmx   1/1     Running   0          2m21s
```

You can stream logs from it:
```
❯ make logs

🔍 Streaming kubernetes-metadata-injector logs...
kubectl logs -l app=kubernetes-metadata-injector -f
time="2021-09-03T04:59:10Z" level=info msg="Listening on port 443..."
time="2021-09-03T05:02:21Z" level=debug msg=healthy uri=/health
```

And hit it's health endpoint from your local machine:
```
❯ curl -k https://localhost:8443/health
OK
```

#### Deploying service
Deploy a valid test service that gets succesfully created:
```
❯ make svc

🚀 Deploying test svc...
kubectl apply -f dev/manifests/svc/injectable-svc.yaml
service/injectable created
```
You should see in the admission webhook logs that the service got mutated .

### Testing

**TODO**

Unit tests can be run with the following command:
```
$ make test
go test ./...
```

### Admission Logic
A set of validations and mutations are implemented in an extensible framework. Those happen on the fly when a service is deployed and no further resources are tracked and updated (ie. no controller logic).


#### Mutating Webhooks
##### Implemented
- [inject annotation](pkg/mutation/service/inject_annotations.go): inject annotation variables taken from configmap into the service based on specific annotation definition

##### How to add a new service mutation
To add a new service mutation, create a file `pkg/mutation/service/MUTATION_NAME.go`, then create a new struct implementing the `service.ServiceMutator` interface.


## Deployment

Deploying the Kubernetes Metadata Injector into your Kubernetes cluster is streamlined through a Helm chart, simplifying the setup process and ensuring that all necessary components and permissions are correctly configured.

### Helm Chart

Located under `helm/metadata-injector`, the provided Helm chart facilitates the deployment of the admission webhook with ease. This chart handles the whole webhook deployment process, including all requisite permissions configuration configurations and certificate handling.

### Chart Installation

```bash
helm repo add metadata-injector https://moveaxlab.github.io/kubernetes-metadata-injector/

helm install --set triggerAnnotationPrefrix='moveax.injector' my-metadata-injector metadata-injector/metadata-injector
```

### Key Features of the Helm Chart

- **Webhook Deployment**: Automatically deploys the Kubernetes Metadata Injector as a Mutating Admission Webhook, ensuring that it intercepts and processes service creation and update requests as configured.
- **Permissions and Configuration**: Sets up all necessary roles, role bindings, and service accounts required for the webhook to function correctly, adhering to the principle of least privilege.
- **MutatingWebhookConfiguration**: Install the `MutatingWebhookConfiguration`, which is crucial for the webhook to correctly intercept and modify requests.
- **Certificate Management**: The deployment process automates certificate generation without the need for a separate cert-manager. This is achieved using the [ingress-nginx/kube-webhook-certgen](https://github.com/kubernetes/ingress-nginx/tree/main/images/kube-webhook-certgen) tool, which runs as a Kubernetes job (helm pre-install hook) to create and manage the necessary certificates.
- **CA Certificate Patching**: Helm post-install hook, the job patches the `MutatingWebhookConfiguration` to include the newly generated CA certificate. This step is essential as it allows the Kubernetes API server to trust the webhook server by validating its certificate against the patched CA certificate. Kubernetes APi server reired mandatory https webserver for admission webhook.
