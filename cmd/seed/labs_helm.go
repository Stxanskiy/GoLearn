package main

// Fixtures + auto-checks for the Helm course. Runs in the k3s golden
// (sandboxImageK8s): helm + kubectl + a running single-node k3s, offline.
//
// TODO: the self-contained labs (local chart create / values / templates /
// upgrade-rollback / hooks) are authorable offline as long as the chart's image
// is a baked one (nginx:alpine, …). The repo-based labs (ch-helm-lab-repos =
// Bitnami charts, ch-helm-lab-helmfile = needs the helmfile binary + remote
// charts) cannot work offline until those are vendored into the golden. Checks
// are being added incrementally; until then Helm lessons keep the manual "Готово".
var helmLabs = map[string]labSpec{}
