# kubernetes-metadata-injector

This is a [Kubernetes admission webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/). It is meant to be used as a validating and mutating admission webhook only and does not support any controller logic. It has been developed as a simple Go web service without using any framework or boilerplate such as kubebuilder.

This project is aimed inject service annotation from secret.

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
Unit tests can be run with the following command:
```
$ make test
go test ./...
?   	github.com/moveaxlab/kubernetes-metadata-injector	[no test files]
ok  	github.com/moveaxlab/kubernetes-metadata-injector/pkg/admission	0.611s
ok  	github.com/moveaxlab/kubernetes-metadata-injector/pkg/mutation	1.064s
```

### Admission Logic
A set of validations and mutations are implemented in an extensible framework. Those happen on the fly when a service is deployed and no further resources are tracked and updated (ie. no controller logic).


#### Mutating Webhooks
##### Implemented
- [inject annotation](pkg/mutation/service/inject_annotations.go): inject annotation variables taken from secret into the service based on specific annotation definition

##### How to add a new service mutation
To add a new service mutation, create a file `pkg/mutation/service/MUTATION_NAME.go`, then create a new struct implementing the `service.ServiceMutator` interface.


## Production Deploy