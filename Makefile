KIND_VERSION = 0.23.0
KUBERNETES_VERSION = 1.31.0

BINDIR = $(CURDIR)/bin
KIND = $(BINDIR)/kind
KUBECTL = $(BINDIR)/kubectl
KIND_CONFIG = $(CURDIR)/kind-config.yaml

.PHONY: build
build:
	CGO_ENABLED=0 go build -o test-sts-client ./cmd/client
	CGO_ENABLED=0 go build -o test-sts-controller ./cmd/controller

.PHONY: run
run:
	./test-sts-controller -kubeconfig ~/.kube/config

.PHONY: start
start: $(KIND) $(KUBECTL)
	$(KIND) create cluster --name=test-sts --config=$(KIND_CONFIG) --image=kindest/node:v$(KUBERNETES_VERSION) --wait 1m
	$(KIND) export kubeconfig --name=test-sts
	$(KUBECTL) create ns sandbox

.PHONY: stop
stop: $(KIND)
	$(KIND) delete cluster --name=test-sts

$(KIND):
	mkdir -p $(BINDIR)
	curl -sfL -o $@ https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-linux-amd64
	chmod a+x $@

$(KUBECTL):
	mkdir -p $(BINDIR)
	curl -sfL -o $@ https://dl.k8s.io/release/v$(KUBERNETES_VERSION)/bin/linux/amd64/kubectl
	chmod a+x $@
