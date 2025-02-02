//nolint:dupl
package keptnevaluation

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	klcv1beta1 "github.com/keptn/lifecycle-toolkit/lifecycle-operator/apis/lifecycle/v1beta1"
	"github.com/keptn/lifecycle-toolkit/lifecycle-operator/apis/lifecycle/v1beta1/common"
	"github.com/keptn/lifecycle-toolkit/lifecycle-operator/controllers/common/config"
	"github.com/keptn/lifecycle-toolkit/lifecycle-operator/controllers/common/eventsender"
	"github.com/keptn/lifecycle-toolkit/lifecycle-operator/controllers/common/telemetry"
	metricsapi "github.com/keptn/lifecycle-toolkit/lifecycle-operator/test/api/metrics/v1beta1"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestKeptnEvaluationReconciler_Reconcile_FailEvaluation(t *testing.T) {

	const namespace = "my-namespace"
	metric := &metricsapi.KeptnMetric{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-metric",
			Namespace: namespace,
		},
		Status: metricsapi.KeptnMetricStatus{
			Value:    "10",
			RawValue: []byte("10"),
		},
	}

	evaluationDefinition := &klcv1beta1.KeptnEvaluationDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-definition",
			Namespace: namespace,
		},
		Spec: klcv1beta1.KeptnEvaluationDefinitionSpec{
			Objectives: []klcv1beta1.Objective{
				{
					KeptnMetricRef: klcv1beta1.KeptnMetricReference{
						Name:      metric.Name,
						Namespace: namespace,
					},
					EvaluationTarget: "<5",
				},
			},
			FailureConditions: klcv1beta1.FailureConditions{
				Retries: 1,
			},
		},
	}

	evaluation := &klcv1beta1.KeptnEvaluation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-evaluation",
			Namespace: namespace,
		},
		Spec: klcv1beta1.KeptnEvaluationSpec{
			EvaluationDefinition: evaluationDefinition.Name,
			FailureConditions: klcv1beta1.FailureConditions{
				Retries: 1,
			},
		},
	}

	reconciler, fakeClient := setupReconcilerAndClient(t, metric, evaluationDefinition, evaluation)

	request := controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      evaluation.Name,
		},
	}

	reconcile, err := reconciler.Reconcile(context.TODO(), request)

	require.Nil(t, err)
	require.True(t, reconcile.Requeue)

	updatedEvaluation := &klcv1beta1.KeptnEvaluation{}
	err = fakeClient.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      evaluation.Name,
	}, updatedEvaluation)

	require.Nil(t, err)

	require.Equal(t, common.StateFailed, updatedEvaluation.Status.EvaluationStatus[metric.Name].Status)
	require.Equal(t, "value '10' did not meet objective '<5'", updatedEvaluation.Status.EvaluationStatus[metric.Name].Message)
}

func TestKeptnEvaluationReconciler_Reconcile_SucceedEvaluation(t *testing.T) {

	const namespace = "my-namespace"
	metric := &metricsapi.KeptnMetric{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-metric",
			Namespace: namespace,
		},
		Status: metricsapi.KeptnMetricStatus{
			Value:    "10",
			RawValue: []byte("10"),
		},
	}

	evaluationDefinition := &klcv1beta1.KeptnEvaluationDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-definition",
			Namespace: namespace,
		},
		Spec: klcv1beta1.KeptnEvaluationDefinitionSpec{
			Objectives: []klcv1beta1.Objective{
				{
					KeptnMetricRef: klcv1beta1.KeptnMetricReference{
						Name:      metric.Name,
						Namespace: namespace,
					},
					EvaluationTarget: "<11",
				},
			},
			FailureConditions: klcv1beta1.FailureConditions{
				Retries: 1,
			},
		},
	}

	evaluation := &klcv1beta1.KeptnEvaluation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-evaluation",
			Namespace: namespace,
		},
		Spec: klcv1beta1.KeptnEvaluationSpec{
			EvaluationDefinition: evaluationDefinition.Name,
			FailureConditions: klcv1beta1.FailureConditions{
				Retries: 1,
			},
		},
	}

	reconciler, fakeClient := setupReconcilerAndClient(t, metric, evaluationDefinition, evaluation)

	request := controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      evaluation.Name,
		},
	}

	reconcile, err := reconciler.Reconcile(context.TODO(), request)

	require.Nil(t, err)
	require.False(t, reconcile.Requeue)

	updatedEvaluation := &klcv1beta1.KeptnEvaluation{}
	err = fakeClient.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      evaluation.Name,
	}, updatedEvaluation)

	require.Nil(t, err)

	require.Equal(t, common.StateSucceeded, updatedEvaluation.Status.EvaluationStatus[metric.Name].Status)
	require.Equal(t, "value '10' met objective '<11'", updatedEvaluation.Status.EvaluationStatus[metric.Name].Message)
}

func TestKeptnEvaluationReconciler_Reconcile_SucceedEvaluation_withDefinitionInDefaultKeptnNamespace(t *testing.T) {

	const namespace = "my-namespace"
	metric := &metricsapi.KeptnMetric{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-metric",
			Namespace: namespace,
		},
		Status: metricsapi.KeptnMetricStatus{
			Value:    "10",
			RawValue: []byte("10"),
		},
	}

	evaluationDefinition := &klcv1beta1.KeptnEvaluationDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-definition",
			Namespace: "keptn",
		},
		Spec: klcv1beta1.KeptnEvaluationDefinitionSpec{
			Objectives: []klcv1beta1.Objective{
				{
					KeptnMetricRef: klcv1beta1.KeptnMetricReference{
						Name:      metric.Name,
						Namespace: namespace,
					},
					EvaluationTarget: "<11",
				},
			},
			FailureConditions: klcv1beta1.FailureConditions{
				Retries: 1,
			},
		},
	}

	evaluation := &klcv1beta1.KeptnEvaluation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-evaluation",
			Namespace: namespace,
		},
		Spec: klcv1beta1.KeptnEvaluationSpec{
			EvaluationDefinition: evaluationDefinition.Name,
			FailureConditions: klcv1beta1.FailureConditions{
				Retries: 1,
			},
		},
	}

	reconciler, fakeClient := setupReconcilerAndClient(t, metric, evaluationDefinition, evaluation)

	request := controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      evaluation.Name,
		},
	}

	reconcile, err := reconciler.Reconcile(context.TODO(), request)

	require.Nil(t, err)
	require.False(t, reconcile.Requeue)

	updatedEvaluation := &klcv1beta1.KeptnEvaluation{}
	err = fakeClient.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      evaluation.Name,
	}, updatedEvaluation)

	require.Nil(t, err)

	require.Equal(t, common.StateSucceeded, updatedEvaluation.Status.EvaluationStatus[metric.Name].Status)
	require.Equal(t, "value '10' met objective '<11'", updatedEvaluation.Status.EvaluationStatus[metric.Name].Message)
}

func setupReconcilerAndClient(t *testing.T, objects ...client.Object) (*KeptnEvaluationReconciler, client.Client) {
	scheme := runtime.NewScheme()

	err := klcv1beta1.AddToScheme(scheme)
	require.Nil(t, err)

	err = metricsapi.AddToScheme(scheme)
	require.Nil(t, err)

	fakeClient := k8sfake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).WithStatusSubresource(objects...).Build()

	provider := metric.NewMeterProvider()
	meter := provider.Meter("keptn/task")

	config.Instance().SetDefaultNamespace("keptn")

	r := &KeptnEvaluationReconciler{
		Client:      fakeClient,
		Scheme:      fakeClient.Scheme(),
		Log:         logr.Logger{},
		EventSender: eventsender.NewK8sSender(record.NewFakeRecorder(100)),
		Meters:      telemetry.SetUpKeptnTaskMeters(meter),
	}
	return r, fakeClient
}
