package main

import (
	"log"
	"strings"

	k8s "github.com/rusik69/govnocloud2/pkg/k8s"
	"github.com/rusik69/govnocloud2/pkg/ssh"
	"github.com/spf13/cobra"
)

// install command
var installCmd = &cobra.Command{
	Use:   "install [master] [workers]",
	Short: "install govnocloud2 cluster",
	Long:  `install govnocloud2 cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("master: ", cfg.Install.Master.Host)
		log.Println("workers ips: ", cfg.Install.Workers.IPs)
		log.Println("workers macs: ", cfg.Install.Workers.MACs)
		log.Println("user: ", cfg.Install.SSH.User)
		log.Println("key: ", cfg.Install.SSH.KeyPath)
		log.Println("dashboard host: ", cfg.Install.Dashboard.Host)
		log.Println("grafana host: ", cfg.Install.Monitoring.GrafanaHost)
		log.Println("prometheus host: ", cfg.Install.Monitoring.PrometheusHost)
		log.Println("alertmanager host: ", cfg.Install.Monitoring.AlertmanagerHost)
		log.Println("kubevirt manager host: ", cfg.Install.Monitoring.KubevirtManagerHost)
		log.Println("longhorn host: ", cfg.Install.Longhorn.Host)
		log.Println("format disk: ", cfg.Install.Longhorn.FormatDisk)
		log.Println("disk: ", cfg.Install.Longhorn.Disk)

		if cfg.Install.Master.Host == "" {
			panic("master is required")
		}

		workersIPsSplit := strings.Split(cfg.Install.Workers.IPs, ",")
		if len(workersIPsSplit) == 0 {
			panic("workers ips are required")
		}

		workersMacsSplit := strings.Split(cfg.Install.Workers.MACs, ",")
		if len(workersMacsSplit) == 0 {
			panic("workers macs are required")
		}

		if len(workersIPsSplit) != len(workersMacsSplit) {
			panic("workers ips and macs should be the same length")
		}

		log.Println("Installing packages on " + cfg.Install.Master.Host)
		out, err := k8s.InstallPackages(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
			"sshpass wakeonlan dnsmasq",
		)
		if err != nil {
			log.Println(out)
			panic(err)
		}

		log.Println("Configuring packages on " + cfg.Install.Master.Host)
		out, err = k8s.ConfigurePackages(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
			cfg.Install.Master.Interface,
			workersMacsSplit,
			workersIPsSplit,
		)
		if err != nil {
			log.Println(out)
			panic(err)
		}

		if cfg.Install.Nat.Enabled {
			log.Println("Setting up NAT")
			err = k8s.SetupNat(
				cfg.Install.Master.Host,
				cfg.Install.SSH.User,
				cfg.Install.SSH.KeyPath,
				cfg.Install.Nat.ExternalInterface,
				cfg.Install.Nat.InternalInterface,
			)
			if err != nil {
				panic(err)
			}
		}

		log.Println("Deploying server on " + cfg.Install.Master.Host)
		err = k8s.Deploy(
			cfg.Install.Master.Host,
			cfg.Server.Host,
			cfg.Web.Host,
			cfg.Server.Port,
			cfg.Web.Port,
			cfg.Install.SSH.User,
			cfg.Install.SSH.Password,
			cfg.Install.SSH.KeyPath,
			cfg.Web.Path,
		)
		if err != nil {
			panic(err)
		}

		// Install k3sup tool
		log.Println("Installing k3sup tool on " + cfg.Install.Master.Host)
		err = k8s.InstallK3sUp(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Deploying k3s master on " + cfg.Install.Master.Host)
		err = k8s.DeployMaster(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			log.Println(out)
			panic(err)
		}
		for _, worker := range workersIPsSplit {
			log.Println("Deploying k3s worker on " + worker)
			err = k8s.DeployNode(
				worker,
				cfg.Install.SSH.User,
				cfg.Install.SSH.KeyPath,
				cfg.Install.SSH.Password,
				cfg.Install.Master.Host,
			)
			if err != nil {
				panic(err)
			}
		}

		command := "sudo k3s kubectl get nodes"
		out, err = ssh.Run(
			command,
			cfg.Install.Master.Host,
			cfg.Install.SSH.KeyPath,
			cfg.Install.SSH.User,
			"",
			false,
			60,
		)
		if err != nil {
			log.Println(out)
			panic(err)
		}
		log.Println("Nodes:")
		log.Println(out)

		log.Println("Installing Etcd")
		err = k8s.InstallEtcd(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}
		// create root user
		log.Println("Creating root user")
		err = k8s.CreateRootUser(cfg.Install.Master.RootPassword)
		if err != nil {
			panic(err)
		}

		log.Println("Installing K9s")
		err = k8s.InstallK9s(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Installing Helm")
		err = k8s.InstallHelm(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Installing Kubernetes Dashboard")
		dashboardToken, err := k8s.InstallDashboard(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
			cfg.Install.Dashboard.Host,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Installing KubeVirt")
		err = k8s.InstallKubeVirt(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
			cfg.Install.Monitoring.KubevirtManagerHost,
			"v1.4.0",
		)
		if err != nil {
			panic(err)
		}

		log.Println("Installing KubeVirt Manager")
		err = k8s.InstallKubeVirtManager(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Installing Longhorn")
		err = k8s.InstallLonghorn(
			cfg.Install.Master.Host,
			workersIPsSplit,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
			cfg.Install.Longhorn.Host,
			cfg.Install.Longhorn.Disk,
			cfg.Install.Longhorn.FormatDisk,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Installing Clickhouse")
		err = k8s.InstallClickhouse(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Installing CNPG")
		err = k8s.InstallCNPG(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Installing MySQL operator")
		err = k8s.InstallMySQL(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}

		log.Println("Installing Ollama")
		err = k8s.InstallOllama(
			cfg.Install.Master.Host,
			cfg.Install.SSH.User,
			cfg.Install.SSH.KeyPath,
		)
		if err != nil {
			panic(err)
		}

		if cfg.Install.Monitoring.Enabled {
			log.Println("Installing monitoring stack")
			err = k8s.DeployPrometheus(
				cfg.Install.Master.Host,
				cfg.Install.SSH.User,
				cfg.Install.SSH.KeyPath,
				cfg.Install.Monitoring.GrafanaHost,
				cfg.Install.Monitoring.PrometheusHost,
				cfg.Install.Monitoring.AlertmanagerHost,
			)
			if err != nil {
				panic(err)
			}
		} else {
			log.Println("Monitoring is not enabled")
		}
		log.Printf("- Dashboard URL: http://%s", cfg.Install.Dashboard.Host)
		log.Printf("- Dashboard Token: %s", dashboardToken)
		log.Printf("- KubeVirt Manager URL: http://%s", cfg.Install.Monitoring.KubevirtManagerHost)
		log.Printf("- Longhorn URL: http://%s", cfg.Install.Longhorn.Host)
		log.Printf("- Prometheus URL: http://%s", cfg.Install.Monitoring.PrometheusHost)
		log.Printf("- Alertmanager URL: http://%s", cfg.Install.Monitoring.AlertmanagerHost)
		log.Printf("- Grafana URL: http://%s", cfg.Install.Monitoring.GrafanaHost)
	},
}
