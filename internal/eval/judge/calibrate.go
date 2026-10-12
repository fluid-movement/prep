package judge

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Sample is one hand-labelled judgement: a clipped session, a criterion and
// the verdict a person gave.
type Sample struct {
	Name       string `yaml:"name"`
	Rubric     string `yaml:"rubric"`
	Task       string `yaml:"task"`
	Transcript string `yaml:"transcript"`
	Diff       string `yaml:"diff"`
	Final      string `yaml:"final"`
	// Verdict is the label: pass or fail.
	Verdict string `yaml:"verdict"`
}

// LoadSamples reads every *.yaml in dir; each file holds one sample or a
// list of them.
func LoadSamples(dir string) ([]Sample, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	var out []Sample
	for _, f := range files {
		data, err := readFile(f)
		if err != nil {
			return nil, err
		}
		var list []Sample
		dec := yaml.NewDecoder(bytes.NewReader(data))
		dec.KnownFields(true)
		if err := dec.Decode(&list); err != nil {
			var one Sample
			d2 := yaml.NewDecoder(bytes.NewReader(data))
			d2.KnownFields(true)
			if err2 := d2.Decode(&one); err2 != nil {
				return nil, fmt.Errorf("%s: %w", f, err)
			}
			list = []Sample{one}
		}
		for _, s := range list {
			if s.Rubric == "" || (s.Verdict != "pass" && s.Verdict != "fail") {
				return nil, fmt.Errorf("%s: sample %q needs a rubric and a verdict of pass or fail", f, s.Name)
			}
			out = append(out, s)
		}
	}
	return out, nil
}

// Report is the outcome of a calibration.
type Report struct {
	// Agreement is the fraction of judgements equal to the human label.
	Agreement float64 `json:"agreement"`
	// Consistency is the fraction of samples on which every repeat agreed
	// with the others.
	Consistency float64 `json:"consistency"`
	Judgements  int     `json:"judgements"`
	Errored     int     `json:"errored"`
	// Disagreements lists sample names the judge got wrong at least once.
	Disagreements []string `json:"disagreements,omitempty"`
	Trusted       bool     `json:"trusted"`
}

// DefaultThreshold is the agreement below which the judge is untrusted.
const DefaultThreshold = 0.85

// Calibrate judges every sample repeats times and compares with the labels.
// Errored judgements count against agreement. The judge's Trusted flag is
// set from the result.
func (j *Judge) Calibrate(ctx context.Context, samples []Sample, repeats int, threshold float64) (Report, error) {
	if len(samples) == 0 {
		return Report{}, fmt.Errorf("no calibration samples")
	}
	if repeats < 1 {
		repeats = 1
	}
	if threshold <= 0 {
		threshold = DefaultThreshold
	}
	var r Report
	agree, consistent := 0, 0
	for _, s := range samples {
		verdicts := map[string]int{}
		wrong := false
		for i := 0; i < repeats; i++ {
			res := j.Grade(ctx, Input{Task: s.Task, Transcript: s.Transcript, Diff: s.Diff, Final: s.Final}, s.Rubric)
			r.Judgements++
			switch {
			case res.Errored:
				r.Errored++
				wrong = true
				verdicts["error"]++
			default:
				v := map[bool]string{true: "pass", false: "fail"}[res.Pass]
				verdicts[v]++
				if v == s.Verdict {
					agree++
				} else {
					wrong = true
				}
			}
		}
		if len(verdicts) == 1 {
			consistent++
		}
		if wrong {
			r.Disagreements = append(r.Disagreements, s.Name)
		}
	}
	r.Agreement = float64(agree) / float64(r.Judgements)
	r.Consistency = float64(consistent) / float64(len(samples))
	r.Trusted = r.Agreement >= threshold
	j.Trusted = &r.Trusted
	return r, nil
}
