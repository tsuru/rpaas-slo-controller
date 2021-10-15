package definition

import (
	"strings"

	"github.com/globocom/slo-generator/methods"
	"github.com/globocom/slo-generator/slo"
	"github.com/tsuru/rpaas-operator/api/v1alpha1"
)

const (
	rpaasTagsAnnotation = "rpaas.extensions.tsuru.io/tags"
)

var classesDefinition = slo.ClassesDefinition{
	Classes: []slo.Class{
		{
			Name: "critical_fast",
			Objectives: slo.Objectives{
				Availability: 99.99,
				Latency: []methods.LatencyTarget{
					{
						LE:     "0.100",
						Target: 99,
					},
					{
						LE:     "0.050",
						Target: 95,
					},
				},
			},
		},
		{
			Name: "critical",
			Objectives: slo.Objectives{
				Availability: 99.99,
				Latency: []methods.LatencyTarget{
					{
						LE:     "0.200",
						Target: 99,
					},
					{
						LE:     "0.100",
						Target: 95,
					},
				},
			},
		},
		{
			Name: "high_fast",
			Objectives: slo.Objectives{
				Availability: 99.9,
				Latency: []methods.LatencyTarget{
					{
						LE:     "0.200",
						Target: 99,
					},
					{
						LE:     "0.100",
						Target: 95,
					},
				},
			},
		},
		{
			Name: "high",
			Objectives: slo.Objectives{
				Availability: 99.9,
				Latency: []methods.LatencyTarget{
					{
						LE:     "1.000",
						Target: 99,
					},
					{
						LE:     "0.500",
						Target: 95,
					},
				},
			},
		},
		{
			Name: "high_slow",
			Objectives: slo.Objectives{
				Availability: 99.9,
				Latency: []methods.LatencyTarget{
					{
						LE:     "5.000",
						Target: 99,
					},
					{
						LE:     "1.000",
						Target: 95,
					},
				},
			},
		},
		{
			Name: "medium",
			Objectives: slo.Objectives{
				Availability: 99,
			},
		},
		{
			Name: "low",
			Objectives: slo.Objectives{
				Availability: 98,
			},
		},
	},
}

func SLOClass(instance *v1alpha1.RpaasInstance) (*slo.Class, error) {
	tagsRaw := instance.ObjectMeta.Annotations[rpaasTagsAnnotation]
	var tags []string
	if tagsRaw != "" {
		tags = strings.Split(tagsRaw, ",")
	}
	sloTags := extractTagValues([]string{"slo:", "SLO:", "slo=", "SLO="}, tags)
	if len(sloTags) == 0 {
		return nil, nil
	}

	class := strings.ToLower(sloTags[0])
	sloClass, err := classesDefinition.FindClass(class)
	if err != nil {
		return nil, err
	}

	return sloClass, nil

}

func extractTagValues(prefixes, tags []string) []string {
	for _, t := range tags {
		for _, p := range prefixes {
			if !strings.HasPrefix(t, p) {
				continue
			}

			separator := string(p[len(p)-1])
			parts := strings.SplitN(t, separator, 2)
			if len(parts) == 1 {
				return nil
			}

			return parts[1:]
		}
	}

	return nil
}
