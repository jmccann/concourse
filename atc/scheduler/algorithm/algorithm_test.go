package algorithm_test

import (
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = DescribeTable("Input resolving",
	(Example).Run,

	Entry("can fan-in", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				// pass a and b
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},

				// pass a but not b
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Passed:   []string{"simple-a", "simple-b"},
			},
		},

		// no v2 as it hasn't passed b
		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv1",
			},
		},
	}),

	Entry("propagates resources together", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 1, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
			},
		},

		Inputs: Inputs{
			{Name: "resource-x", Resource: "resource-x", Passed: []string{"simple-a"}},
			{Name: "resource-y", Resource: "resource-y", Passed: []string{"simple-a"}},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv1",
				"resource-y": "ryv1",
			},
		},
	}),

	Entry("correlates inputs by build, allowing resources to skip jobs", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 1, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},

				{Job: "fan-in", BuildID: 3, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},

				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 4, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
			},
		},

		Inputs: Inputs{
			{Name: "resource-x", Resource: "resource-x", Passed: []string{"simple-a", "fan-in"}},
			{Name: "resource-y", Resource: "resource-y", Passed: []string{"simple-a"}},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv1",

				// not ryv2, as it didn't make it through build relating simple-a to fan-in
				"resource-y": "ryv1",
			},
		},
	}),

	Entry("resolve a resource when it has versions", Example{
		DB: DB{
			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Latest: true},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv1",
			},
		},
	}),

	Entry("does not resolve a resource when it does not have any versions", Example{
		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Pinned: "rxv2"},
			},
		},

		Result: Result{
			OK:     false,
			Values: map[string]string{},
			Errors: map[string]string{
				"resource-x": "pinned version ver:rxv2 not found",
			},
		},
	}),

	Entry("finds only versions that passed through together", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 1, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 2, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},

				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv3", CheckOrder: 2},
				{Job: "simple-a", BuildID: 3, Resource: "resource-y", Version: "ryv3", CheckOrder: 2},
				{Job: "simple-b", BuildID: 4, Resource: "resource-x", Version: "rxv3", CheckOrder: 2},
				{Job: "simple-b", BuildID: 4, Resource: "resource-y", Version: "ryv3", CheckOrder: 2},

				{Job: "simple-a", BuildID: 5, Resource: "resource-x", Version: "rxv2", CheckOrder: 1},
				{Job: "simple-a", BuildID: 5, Resource: "resource-y", Version: "ryv4", CheckOrder: 1},

				{Job: "simple-b", BuildID: 6, Resource: "resource-x", Version: "rxv4", CheckOrder: 1},
				{Job: "simple-b", BuildID: 6, Resource: "resource-y", Version: "rxv4", CheckOrder: 1},

				{Job: "simple-b", BuildID: 7, Resource: "resource-x", Version: "rxv4", CheckOrder: 1},
				{Job: "simple-b", BuildID: 7, Resource: "resource-y", Version: "rxv2", CheckOrder: 1},
			},
		},

		Inputs: Inputs{
			{Name: "resource-x", Resource: "resource-x", Passed: []string{"simple-a", "simple-b"}},
			{Name: "resource-y", Resource: "resource-y", Passed: []string{"simple-a", "simple-b"}},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv3",
				"resource-y": "ryv3",
			},
		},
	}),

	//XXX: revisit unique output names (get and put can currently use the same name, but moving forward we don't want this)
	Entry("can collect distinct versions of resources without correlating by job", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 2, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
			},
		},

		Inputs: Inputs{
			{Name: "simple-a-resource-x", Resource: "resource-x", Passed: []string{"simple-a"}},
			{Name: "simple-b-resource-x", Resource: "resource-x", Passed: []string{"simple-b"}},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"simple-a-resource-x": "rxv1",
				"simple-b-resource-x": "rxv2",
			},
		},
	}),

	Entry("resolves passed constraints with common jobs", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "shared-job", BuildID: 1, Resource: "resource-1", Version: "r1-common-to-shared-and-j1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 1, Resource: "resource-2", Version: "r2-common-to-shared-and-j2", CheckOrder: 1},
				{Job: "job-1", BuildID: 2, Resource: "resource-1", Version: "r1-common-to-shared-and-j1", CheckOrder: 1},
				{Job: "job-2", BuildID: 3, Resource: "resource-2", Version: "r2-common-to-shared-and-j2", CheckOrder: 1},
			},
		},

		Inputs: Inputs{
			{
				Name:     "input-1",
				Resource: "resource-1",
				Passed:   []string{"shared-job", "job-1"},
			},
			{
				Name:     "input-2",
				Resource: "resource-2",
				Passed:   []string{"shared-job", "job-2"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"input-1": "r1-common-to-shared-and-j1",
				"input-2": "r2-common-to-shared-and-j2",
			},
		},
	}),

	Entry("resolves passed constraints with common jobs, skipping versions that are not common to builds of all jobs", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "shared-job", BuildID: 1, Resource: "resource-1", Version: "r1-common-to-shared-and-j1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 1, Resource: "resource-2", Version: "r2-common-to-shared-and-j2", CheckOrder: 1},
				{Job: "job-1", BuildID: 2, Resource: "resource-1", Version: "r1-common-to-shared-and-j1", CheckOrder: 1},
				{Job: "job-2", BuildID: 3, Resource: "resource-2", Version: "r2-common-to-shared-and-j2", CheckOrder: 1},

				{Job: "shared-job", BuildID: 4, Resource: "resource-1", Version: "new-r1-common-to-shared-and-j1", CheckOrder: 2},
				{Job: "shared-job", BuildID: 4, Resource: "resource-2", Version: "new-r2-common-to-shared-and-j2", CheckOrder: 2},
				{Job: "job-1", BuildID: 5, Resource: "resource-1", Version: "new-r1-common-to-shared-and-j1", CheckOrder: 2},
			},
		},

		Inputs: Inputs{
			{
				Name:     "input-1",
				Resource: "resource-1",
				Passed:   []string{"shared-job", "job-1"},
			},
			{
				Name:     "input-2",
				Resource: "resource-2",
				Passed:   []string{"shared-job", "job-2"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"input-1": "r1-common-to-shared-and-j1",
				"input-2": "r2-common-to-shared-and-j2",
			},
		},
	}),

	Entry("finds the latest version for inputs with no passed constraints", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				// build outputs
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 1, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
			},

			Resources: []DBRow{
				// the versions themselves
				// note: normally there's one of these for each version, including ones
				// that appear as outputs
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-y", Version: "ryv4", CheckOrder: 4},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},
				{Resource: "resource-y", Version: "ryv5", CheckOrder: 5},
				{Resource: "resource-x", Version: "rxv5", CheckOrder: 5},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Passed:   []string{"simple-a"},
			},
			{
				Name:     "resource-x-unconstrained",
				Resource: "resource-x",
			},
			{
				Name:     "resource-y-unconstrained",
				Resource: "resource-y",
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x":               "rxv1",
				"resource-x-unconstrained": "rxv5",
				"resource-y-unconstrained": "ryv5",
			},
			// IC: map[string]bool{},
		},
	}),

	Entry("finds the non-disabled latest version for inputs with no passed constraints", Example{
		DB: DB{
			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},
				{Resource: "resource-x", Version: "rxv5", CheckOrder: 5, Disabled: true},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv4",
			},
		},
	}),

	Entry("returns a missing input reason when no input version satisfies the passed constraint", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 2, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Passed:   []string{"simple-a", "simple-b"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Passed:   []string{"simple-a", "simple-b"},
			},
		},

		// only one reason since skipping algorithm if resource does not satisfy passed constraints by itself
		Result: Result{
			OK:     false,
			Values: map[string]string{},
			Errors: map[string]string{
				"resource-x": "passed job 3 does not have a build that satisfies the constraints",
			},
			Skipped: map[string]bool{
				"resource-y": true,
			},
		},
	}),

	Entry("finds next version for inputs that use every version when there is a build for that resource", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 4, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv2",
			},
		},
	}),

	Entry("finds next non-disabled version for inputs that use every version when there is a build for that resource", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 4, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2, Disabled: true},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv3",
			},
		},
	}),

	Entry("finds current non-disabled version if all later versions are disabled for inputs that use every version when there is a build for that resource", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 4, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4, Disabled: true},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv3",
			},
		},
	}),

	Entry("finds last non-disabled version if all later and current versions are disabled for inputs that use every version when there is a build for that resource", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 4, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3, Disabled: true},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4, Disabled: true},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv2",
			},
		},
	}),

	Entry("finds last enabled version for inputs that use every version when there is no builds for that resource", Example{
		DB: DB{
			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},

				{Resource: "resource-z", Version: "rzv1", CheckOrder: 1},
				{Resource: "resource-z", Version: "rzv2", CheckOrder: 2, Disabled: true},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
			},
			{
				Name:     "resource-z",
				Resource: "resource-z",
				Version:  Version{Every: true},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv2",
				"resource-y": "ryv1",
				"resource-z": "rzv1",
			},
		},
	}),

	Entry("finds last version for inputs that use every version when there is no builds for that resource", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 4, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv2",
				"resource-y": "ryv3",
			},
		},
	}),

	Entry("finds next version that passed constraints for inputs that use every version", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv3",
			},
		},
	}),

	Entry("returns the first common version when the current job has no builds and there are multiple passed constraints with version every", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},

				{Job: "simple-b", BuildID: 3, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a", "simple-b"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv1",
			},
		},
	}),

	Entry("does not find candidates when the current job has no builds, there are multiple passed constraints with version every, and a passed job has no builds", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a", "simple-b"},
			},
		},

		Result: Result{
			OK:     false,
			Values: map[string]string{},
			Errors: map[string]string{
				"resource-x": "passed job 2 does not have a build that satisfies the constraints",
			},
		},
	}),

	Entry("returns the next version when there is a passed constraint with version every", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 2, ToBuildID: 1},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv2",
			},
		},
	}),

	Entry("returns current version if there is no version after it that satisifies constraints", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-x", Version: "rxv2", CheckOrder: 1},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 3, ToBuildID: 1},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv2",
			},
		},
	}),

	Entry("returns the common version when there are multiple passed constraints with version every", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 2, ToBuildID: 1},
				{FromBuildID: 5, ToBuildID: 1},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Job: "simple-b", BuildID: 5, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a", "simple-b"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv1",
			},
		},
	}),

	Entry("returns the first version that satisfies constraints when using every version", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-x", Version: "rxv2", CheckOrder: 3},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 3, ToBuildID: 1},
				{FromBuildID: 5, ToBuildID: 1},
			},

			BuildOutputs: []DBRow{
				// only ran for resource-x, not any version of resource-y
				{Job: "shared-job", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},

				// ran for resource-x and resource-y
				{Job: "shared-job", BuildID: 3, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "shared-job", BuildID: 3, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},

				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 5, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"shared-job", "simple-a"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
				Passed:   []string{"shared-job"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv2",
				"resource-y": "ryv1",
			},
		},
	}),

	Entry("does not find candidates when there are multiple passed constraints with version every, and one passed job has no builds", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 2, ToBuildID: 1},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a", "simple-b"},
			},
		},

		Result: Result{
			OK:     false,
			Values: map[string]string{},
			Errors: map[string]string{
				"resource-x": "passed job 3 does not have a build that satisfies the constraints",
			},
		},
	}),

	Entry("returns the latest enabled version when the current job has no builds, and there is a passed constraint with version every", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3, Disabled: true},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv2",
			},
		},
	}),

	Entry("returns the current enabled version when there is a passed constraint with version every, and all later verisons are disabled", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 2, ToBuildID: 1},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2, Disabled: true},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3, Disabled: true},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv1",
			},
		},
	}),

	Entry("returns the latest set of versions that satisfy all passed constraint with version every, and the current job has no builds", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},

				{Job: "simple-b", BuildID: 3, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 4, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},

				{Job: "shared-job", BuildID: 5, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 5, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 6, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 6, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Job: "shared-job", BuildID: 7, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 7, Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a", "shared-job"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
				Passed:   []string{"simple-b", "shared-job"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv1",
				"resource-y": "ryv2",
			},
		},
	}),

	Entry("returns the latest enabled set of versions that satisfy all passed constraint with version every, and the current job has no builds", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},

				{Job: "simple-b", BuildID: 3, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 4, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},

				{Job: "shared-job", BuildID: 5, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 5, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 6, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "shared-job", BuildID: 6, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2, Disabled: true},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a", "shared-job"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
				Passed:   []string{"simple-b", "shared-job"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv1",
				"resource-y": "ryv1",
			},
		},
	}),

	Entry("returns latest build outputs for the passed job that has not run with the current job when using every version", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 100, Resource: "resource-x", Version: "rxv1", CheckOrder: 1}, // resource-y does not have build already
			},

			BuildPipes: []DBRow{
				{FromBuildID: 1, ToBuildID: 100},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv4", CheckOrder: 4},

				{Job: "simple-b", BuildID: 5, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 6, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Job: "simple-b", BuildID: 7, Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
				{Job: "simple-b", BuildID: 8, Resource: "resource-y", Version: "ryv4", CheckOrder: 4},

				{Job: "shared-job", BuildID: 9, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 9, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},

				{Job: "shared-job", BuildID: 10, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 10, Resource: "resource-y", Version: "ryv2", CheckOrder: 1},

				{Job: "shared-job", BuildID: 11, Resource: "resource-x", Version: "rxv2", CheckOrder: 1},
				{Job: "shared-job", BuildID: 11, Resource: "resource-y", Version: "ryv3", CheckOrder: 1},

				{Job: "shared-job", BuildID: 12, Resource: "resource-x", Version: "rxv2", CheckOrder: 1},
				{Job: "shared-job", BuildID: 12, Resource: "resource-y", Version: "ryv4", CheckOrder: 1},

				{Job: "shared-job", BuildID: 13, Resource: "resource-x", Version: "rxv3", CheckOrder: 1},
				{Job: "shared-job", BuildID: 13, Resource: "resource-y", Version: "ryv4", CheckOrder: 1},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
				{Resource: "resource-y", Version: "ryv4", CheckOrder: 4},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a", "shared-job"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
				Passed:   []string{"shared-job", "simple-b"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv2",
				"resource-y": "ryv4",
			},
		},
	}),

	Entry("finds next version that satisfies common constraints when using every version", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 100, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 1, ToBuildID: 100},
				{FromBuildID: 4, ToBuildID: 100},
			},

			BuildOutputs: []DBRow{
				{Job: "shared-job", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},

				{Job: "shared-job", BuildID: 2, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},

				{Job: "shared-job", BuildID: 3, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Job: "shared-job", BuildID: 3, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},

				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 5, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 6, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Job: "simple-b", BuildID: 7, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 8, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Job: "simple-b", BuildID: 9, Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"shared-job", "simple-a"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
				Passed:   []string{"shared-job", "simple-b"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv3",
				"resource-y": "ryv1",
			},
		},
	}),

	Entry("returns the next set of versions that satisfy constraints when using every version", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 100, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: CurrentJobName, BuildID: 100, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 1, ToBuildID: 100},
				{FromBuildID: 5, ToBuildID: 100},
				{FromBuildID: 9, ToBuildID: 100},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv4", CheckOrder: 4},

				{Job: "simple-b", BuildID: 5, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 6, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Job: "simple-b", BuildID: 7, Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
				{Job: "simple-b", BuildID: 8, Resource: "resource-y", Version: "ryv4", CheckOrder: 4},

				{Job: "shared-job", BuildID: 9, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 9, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},

				{Job: "shared-job", BuildID: 10, Resource: "resource-x", Version: "rxv4", CheckOrder: 1},
				{Job: "shared-job", BuildID: 10, Resource: "resource-y", Version: "ryv2", CheckOrder: 1},

				{Job: "shared-job", BuildID: 11, Resource: "resource-x", Version: "rxv4", CheckOrder: 1},
				{Job: "shared-job", BuildID: 11, Resource: "resource-y", Version: "ryv3", CheckOrder: 1},

				{Job: "shared-job", BuildID: 12, Resource: "resource-x", Version: "rxv4", CheckOrder: 1},
				{Job: "shared-job", BuildID: 12, Resource: "resource-y", Version: "ryv4", CheckOrder: 1},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
				{Resource: "resource-y", Version: "ryv4", CheckOrder: 4},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"shared-job", "simple-a"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
				Passed:   []string{"shared-job", "simple-b"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv4",
				"resource-y": "ryv2",
			},
		},
	}),

	Entry("returns earliest set of versions that satisfy the multiple passed constraints with version every when the current job latest build has un-ordered versions", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 2, ToBuildID: 1},
				{FromBuildID: 6, ToBuildID: 1},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Job: "simple-b", BuildID: 5, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 6, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Job: "simple-b", BuildID: 7, Resource: "resource-y", Version: "ryv3", CheckOrder: 3},

				{Job: "shared-job", BuildID: 8, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 8, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 9, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "shared-job", BuildID: 9, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Job: "shared-job", BuildID: 10, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Job: "shared-job", BuildID: 10, Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a", "shared-job"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
				Passed:   []string{"simple-b", "shared-job"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv2",
				"resource-y": "ryv2",
			},
		},
	}),

	Entry("returns earliest set of versions that satisfy the multiple passed constraints with version every when one of the passed jobs skipped a version", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 2, ToBuildID: 1},
				{FromBuildID: 5, ToBuildID: 1},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Job: "simple-b", BuildID: 5, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 7, Resource: "resource-y", Version: "ryv3", CheckOrder: 3},

				{Job: "shared-job", BuildID: 8, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 8, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 9, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "shared-job", BuildID: 9, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Job: "shared-job", BuildID: 10, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Job: "shared-job", BuildID: 10, Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a", "shared-job"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
				Passed:   []string{"simple-b", "shared-job"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv3",
				"resource-y": "ryv3",
			},
		},
	}),

	Entry("returns the current set of versions that satisfy the multiple passed constraints with version every when one of the passed job has no newer versions", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 2, ToBuildID: 1},
				{FromBuildID: 5, ToBuildID: 1},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Job: "simple-b", BuildID: 5, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},

				{Job: "shared-job", BuildID: 8, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 8, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 9, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "shared-job", BuildID: 9, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Job: "shared-job", BuildID: 10, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Job: "shared-job", BuildID: 10, Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a", "shared-job"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
				Passed:   []string{"simple-b", "shared-job"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv1",
				"resource-y": "ryv1",
			},
		},
	}),

	Entry("returns an older set of versions that satisfy the multiple passed constraints with version every when the passed job versions are older than the current set", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-x", Version: "rxv2", CheckOrder: 1},
				{Job: CurrentJobName, BuildID: 1, Resource: "resource-y", Version: "ryv2", CheckOrder: 1},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 2, ToBuildID: 1},
				{FromBuildID: 5, ToBuildID: 1},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Job: "simple-b", BuildID: 5, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 6, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Job: "simple-b", BuildID: 7, Resource: "resource-y", Version: "ryv3", CheckOrder: 3},

				{Job: "shared-job", BuildID: 8, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 8, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"simple-a", "shared-job"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
				Passed:   []string{"simple-b", "shared-job"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv1",
				"resource-y": "ryv1",
			},
		},
	}),

	Entry("returns the earliest non-disabled version that satisfies constraints when several versions do not satisfy when using every version", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 100, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: CurrentJobName, BuildID: 100, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 1, ToBuildID: 100},
				{FromBuildID: 5, ToBuildID: 100},
				{FromBuildID: 9, ToBuildID: 100},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv4", CheckOrder: 4},

				{Job: "simple-b", BuildID: 5, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 6, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Job: "simple-b", BuildID: 7, Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
				{Job: "simple-b", BuildID: 8, Resource: "resource-y", Version: "ryv4", CheckOrder: 4},

				{Job: "shared-job", BuildID: 9, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 9, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},

				{Job: "shared-job", BuildID: 10, Resource: "resource-x", Version: "rxv4", CheckOrder: 1},
				{Job: "shared-job", BuildID: 10, Resource: "resource-y", Version: "ryv2", CheckOrder: 1},

				{Job: "shared-job", BuildID: 11, Resource: "resource-x", Version: "rxv4", CheckOrder: 1},
				{Job: "shared-job", BuildID: 11, Resource: "resource-y", Version: "ryv3", CheckOrder: 1},

				{Job: "shared-job", BuildID: 12, Resource: "resource-x", Version: "rxv4", CheckOrder: 1},
				{Job: "shared-job", BuildID: 12, Resource: "resource-y", Version: "ryv4", CheckOrder: 1},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2, Disabled: true},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
				{Resource: "resource-y", Version: "ryv4", CheckOrder: 4},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"shared-job", "simple-a"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
				Passed:   []string{"shared-job", "simple-b"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv4",
				"resource-y": "ryv3",
			},
		},
	}),

	Entry("returns a missing input reason when no input version satisfies the shared passed constraints", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "shared-job", BuildID: 1, Resource: "resource-1", Version: "r1-common-to-shared-and-j1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 1, Resource: "resource-2", Version: "r2-common-to-shared-and-j2", CheckOrder: 1},

				// resource-1 did not pass job-2 with r1-common-to-shared-and-j1
				{Job: "job-2", BuildID: 3, Resource: "resource-2", Version: "r2-common-to-shared-and-j2", CheckOrder: 1},

				{Job: "shared-job", BuildID: 4, Resource: "resource-1", Version: "new-r1-common-to-shared-and-j1", CheckOrder: 2},
				{Job: "shared-job", BuildID: 4, Resource: "resource-2", Version: "new-r2-common-to-shared-and-j2", CheckOrder: 2},

				// resource-2 did not pass job-1 with new-r2-common-to-shared-and-j2
				{Job: "job-1", BuildID: 5, Resource: "resource-1", Version: "new-r1-common-to-shared-and-j1", CheckOrder: 2},
			},
		},

		Inputs: Inputs{
			{
				Name:     "input-1",
				Resource: "resource-1",
				Passed:   []string{"shared-job", "job-1"},
			},
			{
				Name:     "input-2",
				Resource: "resource-2",
				Passed:   []string{"shared-job", "job-2"},
			},
		},

		Result: Result{
			OK: false,
			Values: map[string]string{
				"input-1": "new-r1-common-to-shared-and-j1",
			},
			Errors: map[string]string{
				"input-2": "passed job 2 does not have a build that satisfies the constraints",
			},
		},
	}),

	Entry("resolves to the pinned version when it exists", Example{
		DB: DB{
			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Pinned: "rxv2"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv2",
			},
		},
	}),

	Entry("does not resolve a version when the pinned version is not in Versions DB (version is disabled or no builds succeeded)", Example{
		DB: DB{
			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				// rxv2 was here
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Pinned: "rxv2"},
			},
		},

		Result: Result{
			OK:     false,
			Values: map[string]string{},
			Errors: map[string]string{
				"resource-x": "pinned version ver:rxv2 not found",
			},
		},
	}),

	// XXX: Passing for the wrong reasons
	Entry("does not resolve a version when the pinned version has not passed the constraint", Example{
		DB: DB{
			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Pinned: "rxv2"},
				Passed:   []string{"some-job"},
			},
		},

		Result: Result{
			OK:     false,
			Values: map[string]string{},
			Errors: map[string]string{"resource-x": "passed job 1 does not have a build that satisfies the constraints"},
		},
	}),

	Entry("resolves to the api pinned version when it exists", Example{
		DB: DB{
			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2, Pinned: true},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv2",
			},
		},
	}),

	Entry("resolves to the pinned version when both api pinned and get step pinned versions exists", Example{
		DB: DB{
			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2, Pinned: true},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Pinned: "rxv1"},
			},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv1",
			},
		},
	}),

	Entry("check orders take precedence over version ID", Example{
		DB: DB{
			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
			},
		},

		Inputs: Inputs{
			{Name: "resource-x", Resource: "resource-x"},
		},

		Result: Result{
			OK: true,
			Values: map[string]string{
				"resource-x": "rxv2",
			},
		},
	}),

	Entry("waiting on upstream job for shared version (ryv3)", Example{
		DB: DB{
			BuildOutputs: []DBRow{
				{Job: "shared-job", BuildID: 1, Resource: "resource-x", Version: "rxv3", CheckOrder: 1},
				{Job: "shared-job", BuildID: 1, Resource: "resource-y", Version: "ryv3", CheckOrder: 1},

				{Job: "simple-a", BuildID: 2, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 3, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Passed:   []string{"shared-job"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Passed:   []string{"shared-job", "simple-a"},
			},
		},

		Result: Result{
			OK: false,
			Values: map[string]string{
				"resource-x": "rxv3",
			},
			Errors: map[string]string{
				"resource-y": "passed job 2 does not have a build that satisfies the constraints",
			},
		},
	}),

	FEntry("reconfigure passed constraints for job with missing upstream dependency (simple-c)", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 100, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: CurrentJobName, BuildID: 100, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: CurrentJobName, BuildID: 100, Resource: "resource-w", Version: "rwv1", CheckOrder: 1},
				{Job: CurrentJobName, BuildID: 100, Resource: "resource-v", Version: "rvv1", CheckOrder: 1},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 1, ToBuildID: 100},
				{FromBuildID: 7, ToBuildID: 100},
				{FromBuildID: 9, ToBuildID: 100},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},

				// {Job: "simple-b", BuildID: 3, Resource: "resource-z", Version: "rzv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 4, Resource: "resource-z", Version: "rzv2", CheckOrder: 2},

				{Job: "simple-c", BuildID: 5, Resource: "resource-w", Version: "rwv3", CheckOrder: 1},
				{Job: "simple-c", BuildID: 6, Resource: "resource-w", Version: "rwv4", CheckOrder: 2},

				{Job: "shared-job", BuildID: 7, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 7, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 8, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "shared-job", BuildID: 8, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},

				{Job: "shared-b", BuildID: 9, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "shared-b", BuildID: 9, Resource: "resource-w", Version: "rwv1", CheckOrder: 1},
				{Job: "shared-b", BuildID: 10, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Job: "shared-b", BuildID: 10, Resource: "resource-w", Version: "rwv2", CheckOrder: 2},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},

				{Resource: "resource-w", Version: "rwv1", CheckOrder: 1},
				{Resource: "resource-w", Version: "rwv2", CheckOrder: 2},
				{Resource: "resource-w", Version: "rwv3", CheckOrder: 3},
				{Resource: "resource-w", Version: "rwv4", CheckOrder: 4},

				{Resource: "resource-z", Version: "rzv1", CheckOrder: 1},
				{Resource: "resource-z", Version: "rzv2", CheckOrder: 2},

				{Resource: "resource-v", Version: "rvv1", CheckOrder: 1},
				{Resource: "resource-v", Version: "rvv2", CheckOrder: 2},
				{Resource: "resource-v", Version: "rvv3", CheckOrder: 3},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Passed:   []string{"shared-job", "simple-a"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Passed:   []string{"shared-job", "shared-b"},
			},
			{
				Name:     "resource-z",
				Resource: "resource-z",
				Passed:   []string{"simple-b"},
			},
			{
				Name:     "resource-w",
				Resource: "resource-w",
				Passed:   []string{"shared-b", "simple-c"},
			},
			{
				Name:     "resource-v",
				Resource: "resource-v",
				Version:  Version{Latest: true},
			},
		},

		Result: Result{
			OK: false,
			Values: map[string]string{
				"resource-x": "rxv2",
				"resource-y": "ryv2",
				"resource-z": "rzv2",
			},
			Errors: map[string]string{
				"resource-w": "passed job 4 does not have a build that satisfies the constraints",
			},
			Skipped: map[string]bool{
				"resource-v": true,
			},
		},
	}),

	Entry("finds a suitable candidate for any inputs resolved before an unresolveable candidates, all candidates after are skipped", Example{
		DB: DB{
			BuildInputs: []DBRow{
				{Job: CurrentJobName, BuildID: 100, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: CurrentJobName, BuildID: 100, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: CurrentJobName, BuildID: 100, Resource: "resource-c", Version: "rcv2", CheckOrder: 2},
			},

			BuildPipes: []DBRow{
				{FromBuildID: 1, ToBuildID: 100},
				{FromBuildID: 9, ToBuildID: 100},
				{FromBuildID: 14, ToBuildID: 100},
			},

			BuildOutputs: []DBRow{
				{Job: "simple-a", BuildID: 1, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "simple-a", BuildID: 2, Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Job: "simple-a", BuildID: 3, Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Job: "simple-a", BuildID: 4, Resource: "resource-x", Version: "rxv4", CheckOrder: 4},

				{Job: "simple-b", BuildID: 6, Resource: "resource-y", Version: "ryv2", CheckOrder: 2},
				{Job: "simple-b", BuildID: 7, Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
				{Job: "simple-b", BuildID: 8, Resource: "resource-y", Version: "ryv4", CheckOrder: 4},
				{Job: "simple-b", BuildID: 6, Resource: "resource-d", Version: "rdv1", CheckOrder: 1},
				{Job: "simple-b", BuildID: 7, Resource: "resource-d", Version: "rdv2", CheckOrder: 2},
				{Job: "simple-b", BuildID: 8, Resource: "resource-d", Version: "rdv4", CheckOrder: 4},

				{Job: "shared-job", BuildID: 9, Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 9, Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Job: "shared-job", BuildID: 9, Resource: "resource-d", Version: "rdv1", CheckOrder: 1},

				{Job: "shared-job", BuildID: 10, Resource: "resource-x", Version: "rxv4", CheckOrder: 1},
				{Job: "shared-job", BuildID: 10, Resource: "resource-y", Version: "ryv2", CheckOrder: 1},

				{Job: "shared-job", BuildID: 11, Resource: "resource-x", Version: "rxv4", CheckOrder: 1},
				{Job: "shared-job", BuildID: 11, Resource: "resource-y", Version: "ryv2", CheckOrder: 1},

				{Job: "shared-job", BuildID: 12, Resource: "resource-x", Version: "rxv4", CheckOrder: 1},
				{Job: "shared-job", BuildID: 12, Resource: "resource-y", Version: "ryv2", CheckOrder: 1},

				{Job: "simple-1", BuildID: 13, Resource: "resource-a", Version: "rav1", CheckOrder: 1},
				{Job: "simple-1", BuildID: 13, Resource: "resource-c", Version: "rcv1", CheckOrder: 1},
				{Job: "simple-1", BuildID: 14, Resource: "resource-a", Version: "rav2", CheckOrder: 2},
				{Job: "simple-1", BuildID: 14, Resource: "resource-c", Version: "rcv2", CheckOrder: 2},
				{Job: "simple-1", BuildID: 15, Resource: "resource-a", Version: "rav3", CheckOrder: 3},
				{Job: "simple-1", BuildID: 15, Resource: "resource-c", Version: "rcv3", CheckOrder: 3},

				{Job: "simple-2", BuildID: 16, Resource: "resource-b", Version: "rbv1", CheckOrder: 1},

				{Job: "simple-3", BuildID: 17, Resource: "resource-d", Version: "rdv1", CheckOrder: 1},
				{Job: "simple-3", BuildID: 18, Resource: "resource-d", Version: "rdv2", CheckOrder: 2},
			},

			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
				{Resource: "resource-x", Version: "rxv3", CheckOrder: 3},
				{Resource: "resource-x", Version: "rxv4", CheckOrder: 4},

				{Resource: "resource-y", Version: "ryv1", CheckOrder: 1},
				{Resource: "resource-y", Version: "ryv2", CheckOrder: 2, Disabled: true},
				{Resource: "resource-y", Version: "ryv3", CheckOrder: 3},
				{Resource: "resource-y", Version: "ryv4", CheckOrder: 4},

				{Resource: "resource-a", Version: "rav1", CheckOrder: 1},
				{Resource: "resource-a", Version: "rav2", CheckOrder: 2},
				{Resource: "resource-a", Version: "rav3", CheckOrder: 3},

				{Resource: "resource-b", Version: "rbv1", CheckOrder: 1},
				{Resource: "resource-b", Version: "rbv2", CheckOrder: 2},
				{Resource: "resource-b", Version: "rbv3", CheckOrder: 3, Disabled: true},

				{Resource: "resource-c", Version: "rcv1", CheckOrder: 1},
				{Resource: "resource-c", Version: "rcv2", CheckOrder: 2},
				{Resource: "resource-c", Version: "rcv3", CheckOrder: 3, Disabled: true},
				{Resource: "resource-c", Version: "rcv4", CheckOrder: 4},

				{Resource: "resource-d", Version: "rdv1", CheckOrder: 1},
				{Resource: "resource-d", Version: "rdv2", CheckOrder: 2},
				{Resource: "resource-d", Version: "rdv3", CheckOrder: 3, Disabled: true},
				{Resource: "resource-d", Version: "rdv4", CheckOrder: 4},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
				Passed:   []string{"shared-job", "simple-a"},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
				Passed:   []string{"shared-job", "simple-b"},
			},
			{
				Name:     "resource-a",
				Resource: "resource-a",
				Version:  Version{Every: true},
				Passed:   []string{"simple-1"},
			},
			{
				Name:     "resource-b",
				Resource: "resource-b",
				Version:  Version{Every: true},
				Passed:   []string{"simple-2"},
			},
			{
				Name:     "resource-c",
				Resource: "resource-c",
				Version:  Version{Every: true},
			},
			{
				Name:     "resource-d",
				Resource: "resource-d",
				Version:  Version{Every: true},
				Passed:   []string{"shared-job", "simple-b", "simple-3"},
			},
		},

		Result: Result{
			OK: false,
			Values: map[string]string{
				"resource-x": "rxv1",
			},
			Errors: map[string]string{
				"resource-y": "passed job 3 does not have a build that satisfies the constraints",
			},
			Skipped: map[string]bool{
				"resource-a": true,
				"resource-b": true,
				"resource-c": true,
				"resource-d": true,
			},
		},
	}),

	Entry("uses partially resolved candidates when there is an error with no passed", Example{
		DB: DB{
			Resources: []DBRow{
				{Resource: "resource-x", Version: "rxv1", CheckOrder: 1},
				{Resource: "resource-x", Version: "rxv2", CheckOrder: 2},
			},
		},

		Inputs: Inputs{
			{
				Name:     "resource-x",
				Resource: "resource-x",
				Version:  Version{Every: true},
			},
			{
				Name:     "resource-y",
				Resource: "resource-y",
				Version:  Version{Every: true},
			},
		},

		Result: Result{
			OK: false,
			Values: map[string]string{
				"resource-x": "rxv2",
			},
			Errors: map[string]string{
				"resource-y": "latest version of resource not found",
			},
		},
	}),
)
