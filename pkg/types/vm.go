package types

// VM is a virtual machine.
type VM struct {
	// Name is the name of the virtual machine.
	Name string `json:"name"`
	// Image is the image of the virtual machine.
	Image string `json:"image"`
	// Size is the size of the virtual machine.
	Size string `json:"size"`
	// Ports is the ports of the virtual machine.
	Ports []VMPort `json:"ports"`
	// Namespace is the namespace of the virtual machine.
	Namespace string `json:"namespace"`
	// Disk is the disk of the virtual machine.
	Disk string `json:"disk"`
	// Status is the status of the virtual machine.
	Status string `json:"status"`
}

// VMPort is a virtual machine port.
type VMPort struct {
	// Name is the name of the virtual machine port.
	Name string `json:"name"`
	// SourcePort is the port of the virtual machine.
	SourcePort int `json:"port"`
	// DestinationPort is the port of the LB.
	DestinationPort int `json:"destinationPort"`
}

// VMSize is a virtual machine size.
type VMSize struct {
	// Name is the name of the virtual machine size.
	Name string `json:"name"`
	// RAM is the RAM of the virtual machine size.
	RAM int `json:"ram"`
	// CPU is the CPU of the virtual machine size.
	CPU int `json:"cpu"`
	// Disk is the disk of the virtual machine size.
	Disk int `json:"disk"`
}

// VMSizes is a map of virtual machine sizes.
var VMSizes = map[string]VMSize{
	"small": VMSize{
		Name: "small",
		RAM:  1024,
		CPU:  1,
		Disk: 10,
	},
	"medium": VMSize{
		Name: "medium",
		RAM:  2048,
		CPU:  2,
		Disk: 20,
	},
	"large": VMSize{
		Name: "large",
		RAM:  4096,
		CPU:  4,
		Disk: 40,
	},
}

// VMDisk is a virtual machine disk.
type VMDisk struct {
	// Name is the name of the virtual machine disk.
	Name string `json:"name"`
	// Size is the size of the virtual machine disk.
	Size int `json:"size"`
}

// VMImage is a virtual machine image.
type VMImage struct {
	// Image is the Image of the virtual machine image.
	Image string `json:"image"`
}

// VMImages is a map of virtual machine images.
var VMImages = map[string]VMImage{
	"ubuntu24": VMImage{
		Image: "quay.io/containerdisks/ubuntu:24.04",
	},
}
