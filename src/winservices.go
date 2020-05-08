package main

import (
	"context"
	"os"

	"github.com/docker/docker/client"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"

	"github.com/newrelic/nri-docker/src/nri"
	"github.com/newrelic/nri-docker/src/raw"
	"github.com/newrelic/nri-docker/src/raw/aws"
)

type argumentList struct {
	Verbose    bool   `default:"false" help:"Print more information to logs."`
	Pretty     bool   `default:"false" help:"Print pretty formatted JSON."`
	NriCluster string `default:"" help:"Optional. Cluster name"`
	HostRoot   string `default:"/host" help:"If the integration is running from a container, the mounted folder pointing to the host root folder"`
	CgroupPath string `default:"" help:"Optional. The path where cgroup is mounted."`
	Fargate    bool   `default:"false" help:"Enables Fargate container metrics fetching. If enabled no metrics are collected from cgroups or Docker. Defaults to false"`
}

const (
	integrationName     = "com.newrelic.docker"
	integrationVersion  = "0.6.0"
	dockerClientVersion = "1.24" // todo: make configurable
)

var (
	args argumentList
)

func main() {
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	log.SetupLogging(args.Verbose)

	var fetcher raw.Fetcher
	var docker nri.DockerClient
	if args.Fargate {
		metadataV3BaseURL, err := aws.MetadataV3BaseURL()
		exitOnErr(err)

		fetcher, err = aws.NewFargateFetcher(metadataV3BaseURL)
		exitOnErr(err)

		docker, err = aws.NewFargateInspector(metadataV3BaseURL)
		exitOnErr(err)
	} else {
		fetcher = raw.NewCGroupsFetcher(
			args.HostRoot,
			args.CgroupPath,
			raw.GetMountsFilePath(),
		)

		var tmpDocker *client.Client
		tmpDocker, err = client.NewEnvClient()
		exitOnErr(err)
		defer tmpDocker.Close()
		tmpDocker.UpdateClientVersion(dockerClientVersion)
		docker = tmpDocker
	}
	sampler, err := nri.NewSampler(fetcher, docker)
	exitOnErr(err)
	exitOnErr(sampler.SampleAll(context.Background(), i))
	exitOnErr(i.Publish())
}

func exitOnErr(err error) {
	if err != nil {
		log.Error(err.Error())
		os.Exit(-1)
	}
}
