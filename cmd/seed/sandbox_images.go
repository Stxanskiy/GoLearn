package main

// What container images each sandbox actually carries.
//
// Lab sandboxes run without network, so an image a lesson names has to be inside
// the sandbox already. Nothing checked that, and it is not a theoretical worry:
// the CKAD probes lab asked for a Pod on redis:alpine, which the Kubernetes VM
// does not have. The Pod sat in ImagePullBackOff forever while the check passed,
// because the check read the Pod's spec and kubectl fills that in whether or not
// the image ever pulls. The student saw a broken Pod and a green tick at once.
//
// These are the sets of the **production** micro-VM root filesystems, which are
// narrower than what the local Docker sandbox preloads. Narrower is what matters:
// production is where students are, so a lesson that works only locally is
// broken. Read with:
//
//	sudo mount -o loop,ro /opt/fc/rootfs-docker.ext4 /mnt/x
//	sudo cat /mnt/x/var/lib/docker/image/overlay2/repositories.json
//	sudo ls /mnt/x/var/lib/rancher/k3s/agent/images/     # for rootfs-k8s
//
// Changing a set here without changing the image is how the drift this catches
// got in. Rebuild the rootfs (deploy/fc-rootfs) and keep the two in step.
var sandboxImages = map[string]map[string]bool{
	// The base image has no container runtime at all, so naming an image in one
	// of its lessons is always a mistake.
	sandboxImage:   {},
	sandboxImagePG: {},

	sandboxImageDocker: set(
		"alpine:latest", "busybox:latest", "golang:1.21-alpine", "hello-world:latest",
		"mysql:8", "nginx:alpine", "nginx:latest", "node:20-alpine",
		"postgres:15", "postgres:15-alpine", "python:3.11-slim", "python:3.12-alpine",
		"redis:7-alpine", "redis:alpine", "ubuntu:22.04",
	),

	// Plus the k3s airgap bundle (pause, coredns, traefik, metrics-server,
	// local-path-provisioner), which lessons never name directly.
	sandboxImageK8s: set(
		"alpine:latest", "busybox:1.28", "busybox:1.36",
		"nginx:1.25-alpine", "nginx:alpine",
	),
}

// knownImageRepos are the repository names a lab may plausibly refer to. The
// scanner needs it to tell an image reference apart from everything else that
// looks like name:value in a shell script — "jsonpath:{...}", "sha256:…",
// "http://host:80".
var knownImageRepos = set(
	"alpine", "busybox", "nginx", "redis", "python", "postgres", "golang",
	"mysql", "node", "ubuntu", "hello-world", "registry", "httpd", "mongo",
	"curlimages/curl", "bitnami/nginx", "nicolaka/netshoot", "traefik",
)

// imageRefExceptions are references that look like images but are not required
// to exist.
var imageRefExceptions = set(
	// A deliberate typo: the CKAD debugging lab asks the student to find why a
	// Pod will not start, and the answer is this misspelled tag.
	"nginx:alpne",
)

func set(items ...string) map[string]bool {
	out := make(map[string]bool, len(items))
	for _, s := range items {
		out[s] = true
	}
	return out
}

// uncheckableTasks are check-typed tasks that deliberately reach the student
// without a checker, keyed "<module>/<lesson>/<task position>".
//
// The bar for adding a line here is a reason the sandbox cannot verify the task
// at all — not "nobody has written the check yet". Everything else is a task
// that silently turned into a "Готово" button.
var uncheckableTasks = set(
	// mount(2) needs CAP_SYS_ADMIN, which lab containers do not get and which
	// the micro-VM does not hand to a lesson either. There is nothing to observe:
	// the mount simply cannot happen, so the student reads about tmpfs and
	// confirms they did.
	"linux-advanced/ch-ladv-lab10/4",
)
