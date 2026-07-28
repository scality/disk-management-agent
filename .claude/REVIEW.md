# Review criteria

Read by the `/review-pr` skill (Scality agent hub) and by anyone reviewing by hand.
Flag problems only — see "What not to flag" at the end.

## What this repo is

`disk-management-agent` is a Kubernetes node agent that discovers physical drives
behind hardware RAID controllers — through `github.com/scality/raidmgmt` — and
publishes them as cluster-scoped `DiscoveredPhysicalDisk` custom resources
(`api/v1alpha1`) for platforms and operators to consume as an inventory. It runs
privileged on the node. Errors use `github.com/scality/go-errors`; tests are testify
for use cases and Ginkgo for the controller, plus `test/e2e`.

## Criteria

| Area | What to check |
|------|---------------|
| Read-only on hardware | Discovery must never mutate controller or drive state: no volume create/delete, no configuration write, no destructive `raidmgmt` call. A new call into the library must be a read operation — this agent's blast radius is physical storage. |
| Privileged surface | The manager already runs with `privileged: true`; any widening (new host mount, new capability, new device access) must be justified in the PR. Prefer the narrowest access that works. |
| Custom resource identity | `DiscoveredPhysicalDisk` names must be deterministic and stable for the same physical drive across restarts, reboots and enumeration order changes — a name derived from an unstable index causes CR churn and duplicate inventory entries. Check for collisions between nodes. |
| Lifecycle & garbage collection | A drive that disappears leads to its CR being removed or marked accordingly; no orphan CRs, no resurrection of a deleted CR, ownership/finalizers consistent with a cluster-scoped resource written by many nodes. |
| Reconcile idempotency & requeue | Reconciling the same object twice changes nothing; requeue delays (e.g. `cacheNotReadyRequeueDelay`) are bounded and never a hot loop; transient errors are returned rather than swallowed into a bare requeue. |
| Error handling | `github.com/scality/go-errors`: failures wrap a package-level sentinel (`errors.Wrap(ErrX, …, errors.CausedBy(rawErr))`) so `errors.Is` keeps matching the category, rather than `fmt.Errorf`/`%w`. No dropped `err`. |
| raidmgmt usage | Pinned to a tagged release in `go.mod`, never a branch. "Not supported by this implementation" must be handled as an expected outcome, not a fatal error. Parsed vendor output is guarded against missing fields and nil dereferences. |
| Partial hardware failure | One controller failing, absent or returning garbage degrades that entry only — the agent keeps publishing what it could discover instead of crash-looping the DaemonSet. |
| Status conventions | Status via the status subresource only, once per reconcile, standard `metav1.Condition` fields. Status and spec fields consumed downstream are a contract. |
| RBAC scoping | `+kubebuilder:rbac` markers match the verbs the code actually uses; `config/rbac` regenerated and committed. |
| CRD compatibility | `v1alpha1` changes stay additive (no removed field, no newly required field, no narrowed validation) unless the PR explains the migration; validation expressed with kubebuilder markers. |
| Generated code in sync | After editing `api/` types or markers, `zz_generated.deepcopy.go` and `config/crd` are regenerated (`make generate manifests`) and committed in the same PR. |
| Context propagation | The reconcile `ctx` is threaded through every client call and every discovery call; long hardware probes respect cancellation instead of blocking a worker. |
| Tests | Use-case logic covered with testify table-driven tests against fakes of the `raidmgmt` boundary; controller behaviour with Ginkgo/envtest; user-visible behaviour with e2e. A new parsing or naming rule needs a test on the produced CR. |
| Docs | Behaviour, CRD or flag changes update `README.md`; conventions update `CONTRIBUTING.md`. |
| Breaking changes | Call out changes to the CR naming scheme, the CRD schema, published fields, flags and env vars — consumers build inventory on them. |

## What not to flag

- Anything the linters already own: `golangci-lint` (`.golangci.yml`), `gofmt`,
  `goimports` — formatting, import order, unused variables, naming.
- Generated files (`zz_generated.*`, `config/crd`) except when they are stale with
  respect to the sources changed in the same PR.
- Markdown or comment wording preferences.
- Refactors unrelated to the PR's purpose.
