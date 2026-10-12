New package `internal/eval/assert/tools`, registering its kinds with the case loader ([Test case format](01M4QKVBSB5FWC25MM7CB8RBNM)); input is the `transcript.Transcript` ([Transcript](01M4QKVBMEVANE3EBQM2V2VKR8)).

- **Interface**: `type Assertion interface { Kind() string; Check(*transcript.Transcript) Result }`, `Result{Pass bool, Reason string, Step int}`. A registry maps kind names to decoders (`Register(kind, func(yaml.Node) (Assertion, error))`) shared with outcome assertions ([Outcome assertions](01M4QKVC87EPF7DT9AN3NTGF6B)) so the case package stays free of kinds.
- **Matching**: commands compare the parsed `Command` (name, flags, project), not raw text; path patterns use `path.Match`/doublestar-style `**` against the tool's file input; `before` uses step indexes of first match.
- **Failure messages**: say what was expected and what was found (`expected guide before edit; first edit at step 4, guide never ran`).
- **Tests**: table tests over small hand-built transcripts for every kind and the command-parsing edge cases.
