package main

// sandboxImage is the container image used by every shell lab. It is built from
// deploy/sandbox/Dockerfile and ships the CLI tools the lessons need, because
// lab containers run with --network none and cannot install anything at runtime.
const sandboxImage = "golearn/sandbox:latest"

// sandboxImagePG is the sandbox image for the SQL course: the base image plus a
// running PostgreSQL server (deploy/sandbox-pg/Dockerfile).
const sandboxImagePG = "golearn/sandbox-pg:latest"

// sandboxImageDocker is the sandbox image for the Docker courses: the base image
// plus a full Docker Engine (deploy/sandbox-docker/Dockerfile). Lab containers
// using it run privileged; the images the lessons need are baked in, since the
// lab has no network.
const sandboxImageDocker = "golearn/sandbox-docker:latest"

// sandboxImageK8s is the sandbox image for the Kubernetes and Helm courses: a
// privileged lab container with a single-node k3s cluster
// (deploy/sandbox-k8s/Dockerfile), images and charts baked in for offline use.
const sandboxImageK8s = "golearn/sandbox-k8s:latest"
