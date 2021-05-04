package setup

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
	"github.com/openshift/addon-operator/e2e"
)

func TestSetup(t *testing.T) {
	ctx := context.Background()
	objs := e2e.LoadObjectsFromDeploymentFiles(t)

	// Create all objects to install the Addon Operator
	for _, obj := range objs {
		err := e2e.Client.Create(ctx, &obj)
		require.NoError(t, err)

		t.Log("created: ", obj.GroupVersionKind().String(),
			obj.GetNamespace()+"/"+obj.GetName())
	}

	t.Run("API available", func(t *testing.T) {
		addonCRD := &apiextensionsv1.CustomResourceDefinition{}
		err := wait.PollImmediate(time.Second, 10*time.Second, func() (done bool, err error) {
			err = e2e.Client.Get(ctx, types.NamespacedName{
				Name: "addons.addons.managed.openshift.io",
			}, addonCRD)
			if err != nil {
				t.Logf("error getting Addons CRD: %v", err)
				return false, nil
			}

			// check CRD Established Condition
			var establishedCond *apiextensionsv1.CustomResourceDefinitionCondition
			for _, c := range addonCRD.Status.Conditions {
				if c.Type == apiextensionsv1.Established {
					establishedCond = &c
					break
				}
			}

			return establishedCond != nil && establishedCond.Status == apiextensionsv1.ConditionTrue, nil
		})
		require.NoError(t, err, "waiting for Addons CRD to be Established")

		// check CRD API
		addonList := &addonsv1alpha1.AddonList{}
		err = e2e.Client.List(ctx, addonList)
		require.NoError(t, err)
	})
}
