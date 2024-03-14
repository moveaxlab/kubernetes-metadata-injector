.PHONY: test
test:
	@echo "\n🛠️  Running unit tests..."
	go test ./...

.PHONY: build
build:
	@echo "\n🔧  Building Go binaries..."
	GOOS=darwin GOARCH=amd64 go build -o bin/admission-webhook-darwin-amd64 .
	GOOS=linux GOARCH=amd64 go build -o bin/admission-webhook-linux-amd64 .

.PHONY: docker-build
docker-build:
	@echo "\n📦 Building kubernetes-metadata-injector Docker image..."
	docker build -t kubernetes-metadata-injector:latest .

# From this point `kind` is required
.PHONY: cluster
cluster:
	@echo "\n🔧 Creating Kubernetes cluster..."
	kind create cluster --config dev/manifests/kind/kind.cluster.yaml

.PHONY: delete-cluster
delete-cluster:
	@echo "\n♻️  Deleting Kubernetes cluster..."
	kind delete cluster

.PHONY: push
push: docker-build
	@echo "\n📦 Pushing admission-webhook image into Kind's Docker daemon..."
	kind load docker-image kubernetes-metadata-injector:latest

.PHONY: deploy-config
deploy-config:
	@echo "\n⚙️  Applying cluster config..."
	kubectl apply -f dev/manifests/cluster-config/

.PHONY: delete-config
delete-config:
	@echo "\n♻️  Deleting Kubernetes cluster config..."
	kubectl delete -f dev/manifests/cluster-config/

.PHONY: deploy
deploy: push delete deploy-config
	@echo "\n🚀 Deploying kubernetes-metadata-injector..."
	kubectl apply -f dev/manifests/webhook/

.PHONY: delete
delete:
	@echo "\n♻️  Deleting kubernetes-metadata-injector deployment if existing..."
	kubectl delete -f dev/manifests/webhook/ || true

.PHONY: cm
cm:
	@echo "\n🚀 Deploying test configMaps..."
	kubectl apply -f dev/manifests/configmaps/

.PHONY: delete-cm
delete-cm:
	@echo "\n♻️ Deleting test configMaps..."
	kubectl delete -f dev/manifests/configmaps/ || true

.PHONY: svc
svc:
	@echo "\n🚀 Deploying test svc..."
	kubectl apply -f dev/manifests/svc/

.PHONY: delete-svc
delete-svc:
	@echo "\n♻️ Deleting test svc..."
	kubectl delete -f dev/manifests/svc/ || true


.PHONY: logs
logs:
	@echo "\n🔍 Streaming kubernetes-metadata-injector logs..."
	kubectl logs -l app=kubernetes-metadata-injector -f

.PHONY: delete-all
delete-all: delete delete-config delete-cm delete-svc
