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
	Ports []int `json:"ports"`
	// Namespace is the namespace of the virtual machine.
	Namespace string `json:"namespace"`
	// Disk is the disk of the virtual machine.
	Disk string `json:"disk"`
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

