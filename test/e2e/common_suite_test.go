// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	envconf "sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

const WAIT_POD_RUNNING_TIMEOUT = time.Second * 180

// doTestCreateSimplePod tests a simple peer-pod can be created.
func doTestCreateSimplePod(t *testing.T, assert CloudAssert) {
	// TODO: generate me.
	namespace := "default"
	name := "simple-peer-pod"
	pod := newPod(namespace, name, "nginx", "kata")

	simplePodFeature := features.New("Simple Peer Pod").
		WithSetup("Create pod", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			if err = client.Resources().Create(ctx, pod); err != nil {
				t.Fatal(err)
			}
			if err = wait.For(conditions.New(client.Resources()).PodRunning(pod), wait.WithTimeout(WAIT_POD_RUNNING_TIMEOUT)); err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("PodVM is created", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			assert.HasPodVM(t, name)

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			if err = client.Resources().Delete(ctx, pod); err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Feature()
	testEnv.Test(t, simplePodFeature)
}
func doTestCreatePodWithEnvVariables(t *testing.T, assert CloudAssert) {
	// TODO: generate me.
	namespace := "default"
	envkey := "STORAGE_KEY"
	envvalue := "true"
	podname := "env-peer-pod"
	pod := newPodWithEnvVar(namespace, podname, envkey, envvalue, "nginx", "kata")

	EnvPodFeature := features.New("Env Peer Pod").
		WithSetup("Create pod", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			if err = client.Resources().Create(ctx, pod); err != nil {
				t.Fatal(err)
			}
			if err = wait.For(conditions.New(client.Resources()).PodRunning(pod), wait.WithTimeout(WAIT_POD_RUNNING_TIMEOUT)); err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("PodVM is created", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			var podlist corev1.PodList
			var stdout, stderr bytes.Buffer
			keycatcher := ""
			if err := cfg.Client().Resources(namespace).List(ctx, &podlist); err != nil {
				t.Fatal(err)
			}
			for _, i := range podlist.Items {
				if i.ObjectMeta.Name == podname {
					if err := cfg.Client().Resources().ExecInPod(ctx, namespace, podname, "nginx", []string{"printenv"}, &stdout, &stderr); err != nil {
						t.Error(stderr.String())
						t.Fatal(err)
					}
				}
			}

			if strings.Contains(stdout.String(), envkey) {
				envarray := strings.Split(stdout.String(), "\n")
				for _, i := range envarray {
					if i == envkey+"="+envvalue {
						logrus.Infof("Environment variable inside the pod is %s", i)
						keycatcher = i
					}
				}

			}
			if keycatcher != envkey+"="+envvalue {
				t.Error("Environment variable not found in the pod")

			}
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			if err = client.Resources().Delete(ctx, pod); err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Feature()
	testEnv.Test(t, EnvPodFeature)
}
