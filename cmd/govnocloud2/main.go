package main

import (
	"log"

	"github.com/rusik69/govnocloud2/pkg/args"
	"github.com/rusik69/govnocloud2/pkg/k3s"
)

func main() {
	arg, err := args.Parse()
	if err != nil {
		panic(err)
	}
	switch arg.Command {
	case "install":
		log.Println("Deploying k3s master on " + arg.Master)
		out, err := k3s.DeployMaster(arg.Master, arg.User, arg.Key)
		if err != nil {
			panic(err)
		}
		log.Println(out)
		token, err := k3s.GetToken(arg.Master, arg.User, arg.Key)
		if err != nil {
			panic(err)
		}
		for _, worker := range arg.Workers {
			log.Println("Deploying k3s worker on " + worker)
			out, err := k3s.DeployNode(worker, arg.User, arg.Key, arg.Master, token)
			if err != nil {
				panic(err)
			}
			log.Println(out)
		}
	case "uninstall":
		log.Println("Uninstalling k3s master on " + arg.Master)
		out, err := k3s.UninstallMaster(arg.Master, arg.User, arg.Key)
		if err != nil {
			panic(err)
		}
		log.Println(out)
		for _, worker := range arg.Workers {
			log.Println("Uninstalling k3s worker on " + worker)
			out, err := k3s.UninstallNode(worker, arg.User, arg.Key)
			if err != nil {
				panic(err)
			}
			log.Println(out)
		}
	}
}
