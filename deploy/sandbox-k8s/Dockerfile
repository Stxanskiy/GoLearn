# Sandbox image for the Kubernetes and Helm courses: the base lab image plus a
# single-node k3s cluster, kubectl and helm. Lab containers using this image run
# privileged (k3s needs it) and keep /var/lib/rancher on a volume.
#
# There is no network at runtime, so the cluster's own images are baked in as
# k3s' airgap tarball and imported on first start.
FROM golearn/sandbox:latest

ENV DEBIAN_FRONTEND=noninteractive

RUN cp /etc/apt/sources.list.d.off/* /etc/apt/sources.list.d/ \
 && apt-get update \
 && apt-get install -y --no-install-recommends iptables conntrack ethtool socat \
 && (cd /etc/apt/sources.list.d && for f in $(ls /etc/apt/sources.list.d.off); do rm -f "$f"; done) \
 && apt-get update \
 && rm -rf /var/lib/apt/lists/partial

COPY bin/k3s  /usr/local/bin/k3s
COPY bin/helm /usr/local/bin/helm
COPY bin/k3s-airgap-images.tar /var/lib/glk8s/airgap-images.tar
# Application images the lessons deploy (nginx, redis, postgres, ...): without
# network the cluster cannot pull them, so they are imported into containerd
# together with k3s' own images.
COPY bin/app-images.tar /var/lib/glk8s/app-images.tar
# Offline stand-in for Docker Hub. Images tagged :latest are always pulled by
# kubelet, so the cluster needs a registry to pull *from* even though every
# image is already on disk.
COPY bin/registry /usr/local/bin/registry
COPY registry /var/lib/glregistry
COPY bin/helmfile /usr/local/bin/helmfile
# Gateway API. k3s brings Traefik but neither the CRDs nor the Gateway provider,
# so both are placed where k3s applies manifests at startup. Doing it here rather
# than in the lesson matters: enabling the provider later means a Helm upgrade and
# a Traefik restart inside a one-vCPU VM, which the lab cannot sensibly wait out.
# This way the cluster is already serving Gateway API by the time a student opens
# the lesson.
COPY bin/gateway-api-crds.yaml /var/lib/rancher/k3s/server/manifests/gateway-api-crds.yaml
COPY traefik-gateway.yaml      /var/lib/rancher/k3s/server/manifests/traefik-gateway.yaml
COPY k8s-start /usr/local/bin/k8s-start
# Offline Helm repository (the `bitnami` repo of the repos lab) plus the script
# that serves it on localhost:8879 — there is no network at runtime.
COPY helm-mirror /var/lib/glhelmrepo
COPY helm-mirror-start /usr/local/bin/helm-mirror-start

RUN chmod +x /usr/local/bin/k3s /usr/local/bin/helm /usr/local/bin/helmfile /usr/local/bin/k8s-start /usr/local/bin/registry /usr/local/bin/helm-mirror-start \
 && ln -s /usr/local/bin/k3s /usr/local/bin/kubectl \
 && ln -s /usr/local/bin/k3s /usr/local/bin/crictl \
 && printf 'export KUBECONFIG=/etc/rancher/k3s/k3s.yaml\n' >> /root/.bashrc \
 && printf 'alias k=kubectl\n' >> /root/.bashrc

ENV KUBECONFIG=/etc/rancher/k3s/k3s.yaml

CMD ["sleep", "infinity"]
