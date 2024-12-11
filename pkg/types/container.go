package types

// Container is a container.
type Container struct {
	// Name is the name of the container.
	Name string `json:"name"`
	// Namespace is the namespace of the container.
	Namespace string `json:"namespace"`
	// Image is the image of the container.
	Image string `json:"image"`
	// Port is the port of the container.
	Port int `json:"port"`
	// CPU is the CPU of the container.
	CPU int `json:"cpu"`
	// RAM is the RAM of the container.
	RAM int `json:"ram"`
	// Disk is the disk of the container.
	Disk int `json:"disk"`
	// Volume is the volume of the container.
	Volume string `json:"volume"`
	// MountPath is the mount path of the container.
	MountPath string `json:"mountPath"`
	// Env is the environment variables of the container.
	Env []string `json:"env"`
}
