package kubernetes

import (
	"testing"

	"github.com/kelda/kelda/blueprint"
	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/minion/kubernetes/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUpdateDeployments(t *testing.T) {
	t.Parallel()
	conn := db.New()
	deploymentsClient := &mocks.DeploymentInterface{}

	// No actions should be taken if we were unable to list the current
	// deployments.
	deploymentsClient.On("List", mock.Anything).Return(nil, assert.AnError).Once()
	updateDeployments(conn, deploymentsClient, nil)
	deploymentsClient.AssertExpectations(t)

	// Test creating a deployment.
	falseRef := false
	securityCtx := &corev1.SecurityContext{
		Privileged: &falseRef,
	}
	annotations := map[string]string{
		dockerfileHashKey: hashStr(""),
		filesHashKey:      hashContainerValueMap(nil),
		envHashKey:        hashContainerValueMap(nil),
		imageKey:          "image",
		"keldaIP":         "ip",
	}
	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hostname",
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						hostnameKey: "hostname",
					},
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Hostname: "hostname",
					Containers: []corev1.Container{
						{
							Name:            "hostname",
							Image:           "image",
							SecurityContext: securityCtx},
					},
					DNSPolicy: corev1.DNSDefault,
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					hostnameKey: "hostname",
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
		},
	}
	conn.Txn(db.ContainerTable, db.BlueprintTable).Run(func(view db.Database) error {
		dbc := view.InsertContainer()
		dbc.Hostname = "hostname"
		dbc.Image = "image"
		dbc.IP = "ip"
		view.Commit(dbc)

		view.InsertBlueprint()
		return nil
	})
	deploymentsClient.On("List", mock.Anything).Return(
		&appsv1.DeploymentList{}, nil).Once()
	deploymentsClient.On("Create", &deployment).Return(nil, nil).Once()
	updateDeployments(conn, deploymentsClient, nil)
	deploymentsClient.AssertExpectations(t)

	// When the deployment already exists, it should be updated.
	newEnv := map[string]blueprint.ContainerValue{
		"key": blueprint.NewString("value"),
	}
	conn.Txn(db.ContainerTable).Run(func(view db.Database) error {
		dbc := view.SelectFromContainer(nil)[0]
		dbc.Env = newEnv
		view.Commit(dbc)
		return nil
	})
	changedDeployment := deployment
	changedDeployment.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{
		{Name: "key", Value: "value"},
	}
	changedDeployment.Spec.Template.Annotations[envHashKey] =
		hashContainerValueMap(newEnv)
	deploymentsClient.On("List", mock.Anything).Return(
		&appsv1.DeploymentList{
			Items: []appsv1.Deployment{deployment},
		}, nil).Once()
	deploymentsClient.On("Update", &changedDeployment).Return(nil, nil).Once()
	updateDeployments(conn, deploymentsClient, nil)
	deploymentsClient.AssertExpectations(t)

	// When a container is removed, its deployment should be removed.
	conn.Txn(db.ContainerTable).Run(func(view db.Database) error {
		view.Remove(view.SelectFromContainer(nil)[0])
		return nil
	})
	deploymentsClient.On("List", mock.Anything).Return(
		&appsv1.DeploymentList{
			Items: []appsv1.Deployment{changedDeployment},
		}, nil).Once()
	deploymentsClient.On("Delete", changedDeployment.Name, mock.Anything).
		Return(nil, nil).Once()
	updateDeployments(conn, deploymentsClient, nil)
	deploymentsClient.AssertExpectations(t)
}

// The pod spec should be exactly the same everytime it's built. Otherwise,
// Kubernetes will think we're creating a different pod, and destroy the
// old one.
func TestMakePodConsistent(t *testing.T) {
	t.Parallel()

	envSecretNameA := "envSecretValueA"
	envSecretNameB := "envSecretValueB"
	fileSecretNameA := "fileSecretValueA"
	fileSecretNameB := "fileSecretValueB"
	sharedSecretName := "sharedSecretValue"

	secretClient := &mocks.SecretClient{}
	secretClient.On("Get", envSecretNameA).Return("envSecretValueA", nil)
	secretClient.On("Get", envSecretNameB).Return("envSecretValueB", nil)
	secretClient.On("Get", fileSecretNameA).Return("fileSecretValueA", nil)
	secretClient.On("Get", fileSecretNameB).Return("fileSecretValueB", nil)
	secretClient.On("Get", sharedSecretName).Return("sharedSecretValue", nil)

	dbc := db.Container{
		Hostname: "hostname",
		Image:    "image",
		FilepathToContent: map[string]blueprint.ContainerValue{
			"a": blueprint.NewString("1"),
			"b": blueprint.NewString("2"),
			"c": blueprint.NewSecret(fileSecretNameA),
			"d": blueprint.NewSecret(fileSecretNameB),
			"e": blueprint.NewSecret(sharedSecretName),
		},
		Env: map[string]blueprint.ContainerValue{
			"a": blueprint.NewString("1"),
			"b": blueprint.NewString("2"),
			"c": blueprint.NewSecret(envSecretNameA),
			"d": blueprint.NewSecret(envSecretNameB),
			"e": blueprint.NewSecret(sharedSecretName),
		},
	}
	pod, ok := makePod(nil, map[string]*corev1.Affinity{}, secretClient, nil, dbc)
	assert.True(t, ok)
	for i := 0; i < 10; i++ {
		newPod, ok := makePod(nil, map[string]*corev1.Affinity{},
			secretClient, nil, dbc)
		assert.True(t, ok)
		assert.Equal(t, pod, newPod)
	}
}

func TestMakePodConfigMap(t *testing.T) {
	t.Parallel()

	fileMap := map[string]string{"foo/bar": "baz"}
	dbc := db.Container{
		FilepathToContent: map[string]blueprint.ContainerValue{
			"foo/bar": blueprint.NewString("baz"),
		},
	}
	pod, ok := makePod(nil, map[string]*corev1.Affinity{}, nil, nil, dbc)
	assert.True(t, ok)
	assert.Len(t, pod.Volumes, 1)
	assert.Len(t, pod.Containers[0].VolumeMounts, 1)

	assert.Equal(t, pod.Volumes[0].ConfigMap.Name, configMapName(fileMap))
	assert.Equal(t, pod.Volumes[0].Name, pod.Containers[0].VolumeMounts[0].Name)
	assert.Equal(t, pod.Containers[0].VolumeMounts[0].MountPath, "foo/bar")
	assert.Equal(t, pod.Containers[0].VolumeMounts[0].SubPath,
		configMapKey("foo/bar"))
}

func TestDeploymentBuilderSecret(t *testing.T) {
	t.Parallel()

	secretClient := &mocks.SecretClient{}

	mySecretName := "mySecret"
	kubeName, _ := secretRef(mySecretName)
	mySecretVal := "mySecretVal"
	containerValueMap := map[string]blueprint.ContainerValue{
		"myKey": blueprint.NewSecret(mySecretName),
	}

	// Test secret whose value isn't set yet.
	secretClient.On("Get", mock.Anything).Return("", assert.AnError).Once()
	dbc := db.Container{
		FilepathToContent: containerValueMap,
	}
	_, ok := makePod(nil, map[string]*corev1.Affinity{}, secretClient, nil, dbc)
	assert.False(t, ok)
	secretClient.AssertExpectations(t)

	// Once the value is set, we should be able to make the pod.
	secretClient.On("Get", mySecretName).Return(mySecretVal, nil).Once()
	pod, ok := makePod(nil, map[string]*corev1.Affinity{}, secretClient, nil, dbc)
	assert.True(t, ok)
	secretClient.AssertExpectations(t)

	assert.Len(t, pod.Volumes, 1)
	assert.Len(t, pod.Containers[0].VolumeMounts, 1)
	assert.Equal(t, pod.Volumes[0].Name, pod.Containers[0].VolumeMounts[0].Name)
	assert.Equal(t, kubeName, pod.Volumes[0].Secret.SecretName)
	assert.Equal(t, "myKey", pod.Containers[0].VolumeMounts[0].MountPath)
	assert.Equal(t, "value", pod.Containers[0].VolumeMounts[0].SubPath)

	secretHashVolume, ok := getSecretEnvHash(pod, mySecretName)
	assert.True(t, ok)

	// Test referencing secrets in environment variables.
	secretClient.On("Get", mySecretName).Return(mySecretVal, nil).Once()
	dbc = db.Container{
		Env: containerValueMap,
	}
	pod, ok = makePod(nil, map[string]*corev1.Affinity{}, secretClient, nil, dbc)
	assert.True(t, ok)
	secretClient.AssertExpectations(t)
	assert.Contains(t, pod.Containers[0].Env, corev1.EnvVar{
		Name: "myKey",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: kubeName,
				},
				Key: "value",
			},
		},
	})

	// The secret hash should be the same whether the secret is referenced from
	// a volume or environment variable.
	secretHashEnv, ok := getSecretEnvHash(pod, mySecretName)
	assert.True(t, ok)
	assert.Equal(t, secretHashVolume, secretHashEnv)

	// If the secret value changes, the container's environment variables
	// should change.
	mySecretVal = "changed"
	secretClient.On("Get", mySecretName).Return(mySecretVal, nil).Once()
	pod, ok = makePod(nil, map[string]*corev1.Affinity{}, secretClient, nil, dbc)
	assert.True(t, ok)
	secretClient.AssertExpectations(t)

	newSecretHash, ok := getSecretEnvHash(pod, mySecretName)
	assert.True(t, ok)
	assert.NotEqual(t, secretHashEnv, newSecretHash)
}

func TestMakePodCustomImage(t *testing.T) {
	t.Parallel()

	readyImage := "readyImage"
	readyDockerfile := "readyDockerfile"
	readyRepoDigest := "readyRepoDigest"

	buildingImage := "buildingImage"
	buildingDockerfile := "buildingDockerfile"

	images := []db.Image{
		{
			Name:       readyImage,
			Dockerfile: readyDockerfile,
			RepoDigest: readyRepoDigest,
			Status:     db.Built,
		},
		{
			Name:       buildingImage,
			Dockerfile: buildingDockerfile,
			Status:     db.Building,
		},
	}

	// Test that a container whose image is pulled from outside the cluster is
	// unchanged.
	regularImage := "alpine"
	dbc := db.Container{Image: regularImage}
	pod, ok := makePod(images, map[string]*corev1.Affinity{}, nil, nil, dbc)
	assert.Equal(t, regularImage, pod.Containers[0].Image)
	assert.True(t, ok)

	// Test that the container whose image is built gets its image rewritten.
	dbc = db.Container{
		Image:      readyImage,
		Dockerfile: readyDockerfile,
	}
	pod, ok = makePod(images, map[string]*corev1.Affinity{}, nil, nil, dbc)
	assert.Equal(t, readyRepoDigest, pod.Containers[0].Image)
	assert.True(t, ok)

	// Test that the cointainer whose image is not ready yet is marked as invalid.
	dbc = db.Container{
		Image:      buildingImage,
		Dockerfile: buildingDockerfile,
	}
	pod, ok = makePod(images, map[string]*corev1.Affinity{}, nil, nil, dbc)
	assert.False(t, ok)
}

func TestMakePodHostPathVolume(t *testing.T) {
	t.Parallel()

	// Test a container that references a known volume.
	volumeName := "volumeName"
	volumeMap := map[string]corev1.Volume{
		volumeName: {Name: volumeName},
	}

	mountPath := "mountPath"
	dbc := db.Container{
		VolumeMounts: []blueprint.VolumeMount{
			{VolumeName: volumeName, MountPath: mountPath},
		},
	}
	pod, ok := makePod(nil, map[string]*corev1.Affinity{}, nil, volumeMap, dbc)
	assert.True(t, ok)
	assert.Equal(t, []corev1.Volume{volumeMap[volumeName]}, pod.Volumes)
	assert.Equal(t, []corev1.VolumeMount{
		{Name: volumeName, MountPath: mountPath},
	}, pod.Containers[0].VolumeMounts)

	// Test a container that references an unknown volume.
	dbc = db.Container{
		VolumeMounts: []blueprint.VolumeMount{
			{VolumeName: "unknown", MountPath: mountPath},
		},
	}
	_, ok = makePod(nil, map[string]*corev1.Affinity{}, nil, volumeMap, dbc)
	assert.False(t, ok)
}

func TestMakeVolume(t *testing.T) {
	t.Parallel()

	exp := corev1.Volume{
		Name: "name",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "path",
			},
		},
	}
	actual, err := makeVolume(blueprint.Volume{
		Name: "name",
		Type: "hostPath",
		Conf: map[string]string{"path": "path"},
	})
	assert.NoError(t, err)
	assert.Equal(t, exp, actual)

	_, err = makeVolume(blueprint.Volume{Type: "unsupported"})
	assert.EqualError(t, err, "unknown volume type: unsupported")
}

func TestMakeDesiredDeploymentsErrors(t *testing.T) {
	t.Parallel()

	// Test that errNoBlueprint is returned when the database doesn't have any
	// blueprints yet.
	conn := db.New()
	_, err := makeDesiredDeployments(conn, nil)
	assert.Equal(t, errNoBlueprint, err)

	// Test that no deployments are returned if the blueprint contains
	// malformed volumes.
	conn.Txn(db.BlueprintTable).Run(func(view db.Database) error {
		bp := view.InsertBlueprint()
		bp.Volumes = []blueprint.Volume{
			{Type: "malformed"},
		}
		view.Commit(bp)
		return nil
	})
	_, err = makeDesiredDeployments(conn, nil)
	assert.EqualError(t, err, "unknown volume type: malformed")
}

func getSecretEnvHash(pod corev1.PodSpec, secretName string) (string, bool) {
	for _, env := range pod.Containers[0].Env {
		if env.Name == "SECRET_HASH_"+secretName {
			return env.Value, true
		}
	}
	return "", false
}
