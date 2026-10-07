- 2026-10-07T06:50:20Z edited by claude-code/opus-5.5: requirement

- 2026-10-07T06:55:08Z edited by claude-code/opus-5.5: depends_on, requirement

- 2026-10-07T07:02:46Z claude-code/opus-5.5: Done: LoadConfig reads config.yaml else config.yaml.dist (ConfigOwn/ConfigDist), Load validates both files under their own names, Apply creates the own config from the dist file, Init writes config.yaml.dist and .prep/.gitignore, record skips ignored paths (gitx.Tracked). README views section explains the split. This repo converted: config.yaml.dist committed, own config.yaml local and ignored. Tests: TestConfigFallsBackToDist, TestConfigRoundTrip (write from dist keeps the header), TestOwnConfigIsNeverStaged.
