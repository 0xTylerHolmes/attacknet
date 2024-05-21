package suite

import (
	"attacknet/cmd/internal/pkg/chaos"
	"time"
)

func ComposeNodeClockSkewTest(description string, targets []*ChaosTargetSelector, skew, duration string, graceDuration *time.Duration) (*chaos.SuiteTest, error) {
	var steps []chaos.PlanStep
	s, err := composeNodeClockSkewPlanSteps(targets, skew, duration)
	if err != nil {
		return nil, err
	}
	steps = append(steps, s...)

	waitStep := composeWaitForFaultCompletionStep()
	steps = append(steps, *waitStep)

	test := &chaos.SuiteTest{
		TestName:  description,
		PlanSteps: steps,
		HealthConfig: chaos.HealthCheckConfig{
			EnableChecks: true,
			GracePeriod:  graceDuration,
		},
	}

	return test, nil
}

func composeNodeRestartTest(description string, targets []*ChaosTargetSelector, graceDuration *time.Duration) (*chaos.SuiteTest, error) {
	var steps []chaos.PlanStep

	s, err := composeNodeRestartSteps(targets)
	if err != nil {
		return nil, err
	}
	steps = append(steps, s...)

	waitStep := composeWaitForFaultCompletionStep()
	steps = append(steps, *waitStep)

	test := &chaos.SuiteTest{
		TestName:  description,
		PlanSteps: steps,
		HealthConfig: chaos.HealthCheckConfig{
			EnableChecks: true,
			GracePeriod:  graceDuration,
		},
	}

	return test, nil
}

func composeIOLatencyTest(description string, targets []*ChaosTargetSelector, delay *time.Duration, percent int, duration *time.Duration, graceDuration *time.Duration) (*chaos.SuiteTest, error) {
	var steps []chaos.PlanStep

	s, err := composeIOLatencySteps(targets, delay, percent, duration)
	if err != nil {
		return nil, err
	}
	steps = append(steps, s...)

	waitStep := composeWaitForFaultCompletionStep()
	steps = append(steps, *waitStep)

	test := &chaos.SuiteTest{
		TestName:  description,
		PlanSteps: steps,
		HealthConfig: chaos.HealthCheckConfig{
			EnableChecks: true,
			GracePeriod:  graceDuration,
		},
	}

	return test, nil
}

func ComposeNetworkLatencyTest(description string, targets []*ChaosTargetSelector, delay, jitter, duration, grace *time.Duration, correlation int) (*chaos.SuiteTest, error) {
	var steps []chaos.PlanStep
	s, err := composeNetworkLatencySteps(targets, delay, jitter, duration, correlation)
	if err != nil {
		return nil, err
	}
	steps = append(steps, s...)

	waitStep := composeWaitForFaultCompletionStep()
	steps = append(steps, *waitStep)

	test := &chaos.SuiteTest{
		TestName:  description,
		PlanSteps: steps,
		HealthConfig: chaos.HealthCheckConfig{
			EnableChecks: true,
			GracePeriod:  grace,
		},
	}

	return test, nil
}

func ComposePacketDropTest(description string, targets []*ChaosTargetSelector, percent int, direction string, duration, grace *time.Duration) (*chaos.SuiteTest, error) {
	var steps []chaos.PlanStep
	s, err := composePacketDropSteps(targets, percent, direction, duration)
	if err != nil {
		return nil, err
	}
	steps = append(steps, s...)

	waitStep := composeWaitForFaultCompletionStep()
	steps = append(steps, *waitStep)

	test := &chaos.SuiteTest{
		TestName:  description,
		PlanSteps: steps,
		HealthConfig: chaos.HealthCheckConfig{
			EnableChecks: true,
			GracePeriod:  grace,
		},
	}

	return test, nil
}
