# Knowledge-policy comparison pilot

This is a bounded, task-specific comparison protocol using the existing
`internal/eval/fixture` builder and `eval/fixtures/webapp.yaml`; it adds no
production runner or alternate knowledge format.

Pinned source: `bacc2245668c64831389dced173514b2db6000df`. The builder uses the
existing `natural-fix-matching-issue` fixture (same code, issue state, seed,
and deterministic clock) and changes only `.prep/knowledge` content:

| Condition | Knowledge at task start |
| --- | --- |
| `code-only` | The shared `/overview.md` scaffold only; no advisory component entries. |
| `current` | All existing `webapp.yaml` entries, including profile fields, plus the shared scaffold. |
| `lean` | The shared scaffold and one short profile-fields entry retaining constraint, decision, and navigation hints. |
| `stale` | The shared scaffold and one profile-fields entry falsely claiming a 200-character database limit and no rendering truncation. |

Every workspace also receives the same Codex prep skill copied from
`plugins/codex/skills/prep/SKILL.md`; only the advisory project knowledge
entries vary. Each trial's `prep` wrapper applies `PREP_ACTOR` and
`PREP_SESSION` to individual prep invocations, not to the Codex CLI process.

Run three independent Codex CLI sessions in fresh copies of each condition
(12 total). Keep the prompt in `task.md` byte-identical. For each session,
save the complete JSONL event stream only in a private temporary results
directory, plus the CLI's model metadata/version. Record the elapsed wall
time, any token counters supplied by the CLI, and the resulting repo diff and
`prep show`/review outcomes. Produce only an aggregate, sanitized summary in
the issue findings; do not retain raw session content in the repository.

Primary result: contract compliance, checked from the fixture state and CLI
records (tracked issue used, implementation change, no unsupported claims,
required knowledge report/completion gate). Exploratory results: behavior
correctness against the stated criterion, missed constraints, run time and
token counts where available, and maintenance burden measured as knowledge
files/lines inspected or changed. Report each cell's `n`, all failures, and
limits; three repetitions are a pilot, not evidence of superiority. Do not
compare costs across providers/models.

Codex CLI version and model must be recorded from each actual run. This pilot
does not use Claude: live Claude runs remain unauthorized in
`/components/eval-cases.md#tracking-policy-and-deferred-model-checks`.
Existing `prep eval` matrices, variants, fixture conventions, and result
reports remain the canonical infrastructure; this one-task pilot reuses its
fixture API because the current live runner is Claude-specific and its
variants overlay plugin files, not project knowledge.

## Reproduction

From the repository root, make a clean source extraction and overlay this
experiment's fixture builder (the research artifact is newer than the pinned
production snapshot):

```sh
git archive --format=tar -o /private/tmp/prep-kp-base.tar bacc2245668c64831389dced173514b2db6000df
mkdir -p /private/tmp/prep-kp-base
tar -xf /private/tmp/prep-kp-base.tar -C /private/tmp/prep-kp-base
mkdir -p /private/tmp/prep-kp-base/eval/experiments/knowledge-policy
cp eval/experiments/knowledge-policy/build-fixture.go /private/tmp/prep-kp-base/eval/experiments/knowledge-policy/
GOCACHE=/private/tmp/prep-kp-gocache go -C /private/tmp/prep-kp-base build -o /private/tmp/prep-kp-bin ./cmd/prep
git show 65067e922a4962c7a8d44863843c1ac8a62e43ff:plugins/codex/skills/prep/SKILL.md > /private/tmp/prep-kp-skill.md
python3 eval/experiments/knowledge-policy/run-pilot.py \
  --source /private/tmp/prep-kp-base \
  --prep-binary /private/tmp/prep-kp-bin \
  --plugin-skill /private/tmp/prep-kp-skill.md \
  --skill-revision 65067e922a4962c7a8d44863843c1ac8a62e43ff \
  --model gpt-6.1-sol \
  --out /private/tmp/prep-kp-results --repetitions 3
```

The runner stores full JSONL streams only under `--out`; its JSON summary
records requested model separately from any model name the CLI emits. Each
fixture gets the same seed and issue ID, and the CLI `show` output for every
starting knowledge entry is hashed via the CLI into the per-run manifest. It
also retains exact standalone verification command text, outer command status,
and a short command output excerpt. It does not infer subcommand status from a
compound shell command.

The old static proxy is retained only as an obsolete measurement. For the
post-hoc helper check, use a Node version that can import TypeScript and
record `process.version`. In each completed fixture workspace:

```sh
node --experimental-strip-types --input-type=module -e 'import {profileHeadline} from "./src/lib/profile.ts"; const value="x".repeat(200); if(profileHeadline({headline:value})!==value) process.exit(1); console.log(process.version, "200-character helper check passed");'
```

Run the same check against the original fixture as a negative control. This
checks the helper boundary, without browser or database verification. The
findings preserve the testing-constraint confound and recommend a corrected
lean overlay; they do not establish a usefulness ranking.
