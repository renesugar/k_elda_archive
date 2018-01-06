package kubernetes

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"sort"

	"github.com/kelda/kelda/blueprint"
	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/join"
	"github.com/kelda/kelda/util/str"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsclient "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/util/retry"
)

// The default mode to create files specified through Container.filepathToContent.
var filepathToContentMode = int32(0644)

var errNoBlueprint = errors.New("no blueprint in database")

// updateDeployments syncs the containers specified by the user into Kubernetes
// deployments.
// If a deployment already exists with the same name, but possibly different
// values, we just call the "Update" endpoint, and let Kubernetes handle
// figuring out whether the deployment changed or not. This way, we don't have
// to figure out whether some subtle attributes, such as affinities, have
// changed. The downside of this is that the deployments must be exactly the
// same in order to prevent Kubernetes from erroneously restarting containers.
// Therefore, many of the deployment fields are sorted to ensure consistency.
func updateDeployments(conn db.Conn, deploymentsClient appsclient.DeploymentInterface,
	secretClient SecretClient) {

	currentDeployments, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("Failed to list current deployments")
		return
	}

	desiredDeployments, err := makeDesiredDeployments(conn, secretClient)
	if err != nil {
		if err == errNoBlueprint {
			// It's expected for there to not be a blueprint when the machine
			// first boots, and the foreman hasn't connected yet.
			return
		}
		log.WithError(err).Error("Failed to create desired deployments")
	}

	key := func(intf interface{}) interface{} {
		return intf.(appsv1.Deployment).Name
	}
	pairs, toCreate, toDelete := join.HashJoin(
		deploymentSlice(desiredDeployments),
		deploymentSlice(currentDeployments.Items),
		key, key)

	for _, pair := range pairs {
		// Retry updating the deployment if the apiserver reports that there's
		// a conflict. Conflicts are benign -- for example, there might be a
		// conflict if Kubernetes updated the deployment to change the pod
		// status.
		deployment := pair.L.(appsv1.Deployment)
		c.Inc("Update deployment")
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			_, err := deploymentsClient.Update(&deployment)
			return err
		})
		if err != nil {
			log.WithError(err).WithField("deployment", deployment.Name).
				Error("Failed to update deployment")
		}
	}

	for _, intf := range toCreate {
		deployment := intf.(appsv1.Deployment)
		log.WithField("deployment", deployment.Name).Info("Creating deployment")
		c.Inc("Create deployment")
		if _, err := deploymentsClient.Create(&deployment); err != nil {
			log.WithError(err).WithField("deployment", deployment.Name).
				Error("Failed to create deployment")
		}
	}

	for _, intf := range toDelete {
		deployment := intf.(appsv1.Deployment)
		log.WithField("deployment", deployment.Name).Info("Deleting deployment")
		c.Inc("Delete deployment")
		err := deploymentsClient.Delete(deployment.Name, &metav1.DeleteOptions{})
		if err != nil {
			log.WithError(err).WithField("deployment", deployment.Name).
				Error("Failed to delete deployment")
		}
	}
}

const (
	hostnameKey       = "hostname"
	envHashKey        = "env-hash"
	filesHashKey      = "files-hash"
	dockerfileHashKey = "dockerfile-hash"
	imageKey          = "friendly-image"
)

func makeDesiredDeployments(conn db.Conn, secretClient SecretClient) (
	[]appsv1.Deployment, error) {

	var containers []db.Container
	var images []db.Image
	var idToAffinity map[string]*corev1.Affinity
	var volumes []blueprint.Volume
	tables := []db.TableType{db.ContainerTable, db.ImageTable, db.PlacementTable,
		db.BlueprintTable}
	err := conn.Txn(tables...).Run(func(view db.Database) error {
		bp, err := view.GetBlueprint()
		if err != nil {
			return err
		}

		containers = view.SelectFromContainer(func(dbc db.Container) bool {
			return dbc.IP != ""
		})
		images = view.SelectFromImage(nil)
		idToAffinity = toAffinities(view.SelectFromPlacement(nil))
		volumes = bp.Volumes
		return nil
	})
	if err != nil {
		return nil, errNoBlueprint
	}

	volumeMap := map[string]corev1.Volume{}
	for _, volume := range volumes {
		volumeMap[volume.Name], err = makeVolume(volume)
		if err != nil {
			return nil, err
		}
	}

	var deployments []appsv1.Deployment
	for _, dbc := range containers {
		pod, ok := makePod(images, idToAffinity, secretClient, volumeMap, dbc)
		if ok {
			deployments = append(deployments, makeDeployment(dbc, pod))
		}
	}
	return deployments, nil
}

func makeDeployment(dbc db.Container, pod corev1.PodSpec) appsv1.Deployment {
	// These annotations are used by the join in `updateStatuses` to match
	// up Kubernetes pods with the containers in the database.
	annotations := map[string]string{
		dockerfileHashKey: hashStr(dbc.Dockerfile),
		filesHashKey:      hashContainerValueMap(dbc.FilepathToContent),
		envHashKey:        hashContainerValueMap(dbc.Env),
		imageKey:          dbc.Image,
		"keldaIP":         dbc.IP,
	}
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: dbc.Hostname,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						hostnameKey: dbc.Hostname,
					},
					Annotations: annotations,
				},
				Spec: pod,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					hostnameKey: dbc.Hostname,
				},
			},
			// Roll out pods by destroying the previous ones before creating
			// the new ones, rather than trying to create the new pod version
			// before destroying the old one. This way, there are never two
			// pods with the same keldaIP, which can cause issues for the CNI
			// plugin.
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
		},
	}
}

// makePod returns a Kubernetes representation of the given containers as a pod.
// The returned boolean indicates whether it's possible to create a pod spec at
// this time. It's not necessarily an error if a pod can't be created -- for
// example, the user might need to run `kelda secret`, or images might still be
// building.
func makePod(images []db.Image, idToAffinity map[string]*corev1.Affinity,
	secretClient SecretClient, volumeMap map[string]corev1.Volume,
	dbc db.Container) (corev1.PodSpec, bool) {

	// If the container isn't built by Kelda, the image doesn't have to be
	// rewritten to the version hosted by the local registry.
	image := dbc.Image
	if dbc.Dockerfile != "" {
		var foundImage bool
		for _, img := range images {
			if img.Name == dbc.Image && img.Dockerfile == dbc.Dockerfile &&
				img.Status == db.Built {
				foundImage = true
				image = img.RepoDigest
				break
			}
		}
		if !foundImage {
			return corev1.PodSpec{}, false
		}
	}

	env, missing := makeSecretHashEnvVars(secretClient, dbc.GetReferencedSecrets())
	if len(missing) != 0 {
		return corev1.PodSpec{}, false
	}
	env = append(env, makePodEnvVars(dbc.Env)...)

	volumes, volumeMounts := makeVolumesForFilepathToContent(dbc.FilepathToContent)
	for _, volumeMount := range dbc.VolumeMounts {
		volume, ok := volumeMap[volumeMount.VolumeName]
		if !ok {
			log.WithField("volume", volumeMount.VolumeName).
				WithField("container", dbc.Hostname).
				Warn("Unknown volume reference")
			return corev1.PodSpec{}, false
		}

		volumes = append(volumes, volume)
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			MountPath: volumeMount.MountPath,
			Name:      volumeMount.VolumeName,
		})
	}

	// Sort the volumes and volume mounts so that the pod config is
	// consistent. Otherwise, Kubernetes will treat differences in
	// orderings as a reason to restart the pod.
	sort.Sort(volumeMountSlice(volumeMounts))
	sort.Sort(volumeSlice(volumes))
	sort.Sort(envVarSlice(env))

	return corev1.PodSpec{
		Hostname: dbc.Hostname,
		Containers: []corev1.Container{
			{
				Name:         dbc.Hostname,
				Image:        image,
				Env:          env,
				Args:         dbc.Command,
				VolumeMounts: volumeMounts,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &dbc.Privileged,
				},
			},
		},
		Affinity:  idToAffinity[dbc.Hostname],
		DNSPolicy: corev1.DNSDefault,
		Volumes:   volumes,
	}, true
}

// makeSecretHashEnvVars creates environment variables that represent the value
// of the secrets referenced by the container. This way, if a secret value
// changes, these environment variables will change, and Kubernetes will
// restart the container and pick up the new secret value.
func makeSecretHashEnvVars(secretClient SecretClient, secretNames []string) (
	envVars []corev1.EnvVar, missing []string) {
	for _, name := range secretNames {
		val, err := secretClient.Get(name)
		if err != nil {
			missing = append(missing, name)
			continue
		}

		envVars = append(envVars, corev1.EnvVar{
			Name:  "SECRET_HASH_" + name,
			Value: fmt.Sprintf("%x", hashStr(val)),
		})
	}
	return
}

func makePodEnvVars(dbcEnv map[string]blueprint.ContainerValue) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	rawStrings, secrets := blueprint.DivideContainerValues(dbcEnv)
	for key, val := range rawStrings {
		envVars = append(envVars, corev1.EnvVar{
			Name:  key,
			Value: val,
		})
	}

	for key, secret := range secrets {
		kubeName, subpath := secretRef(secret)
		envVars = append(envVars, corev1.EnvVar{
			Name: key,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: kubeName,
					},
					Key: subpath,
				},
			},
		})
	}
	return envVars
}

// makeVolumesForFilepathToContent returns the appropriate Volumes and
// VolumeMounts such that the given filepathToContent will be reflected in the
// container.
// For string files it uses ConfigMap volumes (whose contents are created in
// configmap.go), and for secrets it uses Secret volumes.
func makeVolumesForFilepathToContent(
	filepathToContent map[string]blueprint.ContainerValue) (
	volumes []corev1.Volume, mounts []corev1.VolumeMount) {

	rawStrings, secrets := blueprint.DivideContainerValues(filepathToContent)
	mountedSecretVolumes := map[string]struct{}{}
	for path, secret := range secrets {
		kubeName, key := secretRef(secret)
		volumeName := "secret-volume-" + kubeName

		// If there are multiple references to the same secret, only mount its
		// secret volume once to avoid two references to the exact same volume.
		if _, ok := mountedSecretVolumes[volumeName]; !ok {
			volumes = append(volumes, corev1.Volume{
				Name: volumeName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: kubeName,
					},
				},
			})
			mountedSecretVolumes[volumeName] = struct{}{}
		}

		mounts = append(mounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: path,
			SubPath:   key,
		})
	}

	// Mount the raw string values by mounting the ConfigMap corresponding to
	// the filepathToContent.
	const filesVolumeName = "filepath-to-content"
	if len(rawStrings) != 0 {
		volumes = append(volumes, corev1.Volume{
			Name: filesVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName(rawStrings),
					},
					DefaultMode: &filepathToContentMode,
				},
			},
		})
	}

	for path := range rawStrings {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      filesVolumeName,
			MountPath: path,
			SubPath:   configMapKey(path),
		})
	}
	return
}

func makeVolume(volume blueprint.Volume) (corev1.Volume, error) {
	kubeVolume := corev1.Volume{
		Name: volume.Name,
	}
	switch volume.Type {
	case "hostPath":
		kubeVolume.HostPath = &corev1.HostPathVolumeSource{
			Path: volume.Conf["path"],
		}
	default:
		return corev1.Volume{}, fmt.Errorf("unknown volume type: %s", volume.Type)
	}
	return kubeVolume, nil
}

func hashStr(toHash string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(toHash)))
}

func hashContainerValueMap(containerValMap map[string]blueprint.ContainerValue) string {
	strValMap := map[string]string{}
	for k, v := range containerValMap {
		strValMap[k] = v.String()
	}
	return hashStr(str.MapAsString(strValMap))
}

type deploymentSlice []appsv1.Deployment

func (slc deploymentSlice) Get(ii int) interface{} {
	return slc[ii]
}

func (slc deploymentSlice) Len() int {
	return len(slc)
}

type volumeMountSlice []corev1.VolumeMount

func (slc volumeMountSlice) Len() int {
	return len(slc)
}

func (slc volumeMountSlice) Swap(i, j int) {
	slc[i], slc[j] = slc[j], slc[i]
}

func (slc volumeMountSlice) Less(i, j int) bool {
	return slc[i].MountPath < slc[j].MountPath
}

type volumeSlice []corev1.Volume

func (slc volumeSlice) Len() int {
	return len(slc)
}

func (slc volumeSlice) Swap(i, j int) {
	slc[i], slc[j] = slc[j], slc[i]
}

func (slc volumeSlice) Less(i, j int) bool {
	return slc[i].Name < slc[j].Name
}

type envVarSlice []corev1.EnvVar

func (slc envVarSlice) Len() int {
	return len(slc)
}

func (slc envVarSlice) Swap(i, j int) {
	slc[i], slc[j] = slc[j], slc[i]
}

func (slc envVarSlice) Less(i, j int) bool {
	return slc[i].Name < slc[j].Name
}
