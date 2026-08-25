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
COPY k8s-start /usr/local/bin/k8s-start

RUN chmod +x /usr/local/bin/k3s /usr/local/bin/helm /usr/local/bin/k8s-start /usr/local/bin/registry \
 && ln -s /usr/local/bin/k3s /usr/local/bin/kubectl \
 && ln -s /usr/local/bin/k3s /usr/local/bin/crictl \
 && printf 'export KUBECONFIG=/etc/rancher/k3s/k3s.yaml\n' >> /root/.bashrc \
 && printf 'alias k=kubectl\n' >> /root/.bashrc

ENV KUBECONFIG=/etc/rancher/k3s/k3s.yaml

CMD ["sleep", "infinity"]
