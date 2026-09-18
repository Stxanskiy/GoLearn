# Kubernetes manifests

Reference copies of what runs in k3s. The live cluster is reconciled by ArgoCD from
`Stxanskiy/vibiay_manifests` (path `golearn`) — see `deploy/berg/README.md`. Changes here
have to be copied there to take effect.

## Object storage (MinIO)

Uploaded images — course and specialization icons, covers, lesson pictures — live in an
S3 bucket instead of the database. `minio.yaml` runs a single-node MinIO with its own PVC,
and the app talks to it over the cluster service `minio:9000`.

```bash
kubectl -n golearn create secret generic golearn-secrets \
  --from-literal=S3_ACCESS_KEY=... --from-literal=S3_SECRET_KEY=... \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f deploy/k8s/minio.yaml
```

The bucket is created on first start and gets a public read policy, so the browser loads
images straight from storage. `S3_PUBLIC_URL` is the base URL it uses.

**Pick the public host deliberately.** A separate host (`https://media.prod-factory.ru/golearn`)
keeps uploaded files off the application origin, which is the safer default. Serving them
from a path of the main host works too — uploaded SVG is stored with
`Content-Disposition: attachment`, so opening one directly downloads it instead of running
its scripts — but then an upload shares the origin with the app, and that is worth avoiding
when the extra DNS name is cheap.

Either way the chosen URL must reach `minio:9000` with the bucket in the path, and
`S3_PUBLIC_URL` in `config.yaml` must match it exactly: the app stores absolute URLs in the
database and recognises its own objects by that prefix when replacing or deleting them.

Without `S3_ENDPOINT` the server still runs: icon and lesson-image uploads answer
`503 storage_disabled`, and covers fall back to inline data URIs, as before.
