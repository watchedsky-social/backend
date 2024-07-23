package main

import (
	"context"
	"fmt"
	"log"

	"github.com/watchedsky-social/backend/ci/internal/dagger"
)

type Backend struct{}

func (b *Backend) BuildEnv(
	source *dagger.Directory,
	//+default="registry.lab.verysmart.house"
	registry string,
	//+optional
	//+default="latest"
	imageVersion string,
) *dagger.Container {
	site := dag.Frontend().
		WithRegistry(registry).
		WithImageVersion(imageVersion).
		GetBuiltSite()

	return dag.Container().
		From("cgr.dev/chainguard/go:latest").
		WithDirectory("/src", source, dagger.ContainerWithDirectoryOpts{
			Exclude: []string{"frontend/**"},
		}).
		WithDirectory("/src/frontend", site).
		WithWorkdir("/src").
		WithExec([]string{"go", "mod", "download"})
}

func (b *Backend) Build(
	source *dagger.Directory,
	//+default="registry.lab.verysmart.house"
	registry string,
	//+optional
	//+default="0.0.0-local"
	appVersion string,
	//+optional
	//+default="latest"
	imageVersion string,
) *dagger.Container {
	builtDirectory := b.BuildEnv(source, registry, imageVersion).
		WithExec([]string{"mkdir", "/assets"}).
		WithExec([]string{"go", "build", "-o", "/assets/server",
			"-ldflags", fmt.Sprintf("-X main.Version=%s", appVersion), "./cmd/server"}).
		Directory("/assets")

	return dag.Container().
		From("cgr.dev/chainguard/glibc-dynamic:latest").
		WithDirectory("/assets", builtDirectory).
		WithExposedPort(8000).
		WithEntrypoint([]string{"/assets/server"})
}

func (b *Backend) BuildAndPublish(
	ctx context.Context,
	source *dagger.Directory,
	//+default="registry.lab.verysmart.house"
	registry string,
	//+optional
	//+default="0.0.0-local"
	appVersion string,
	//+optional
	//+default="latest"
	imageVersion string,
	username string,
	password *dagger.Secret,
) ([]string, error) {
	builtContainer := b.Build(source, registry, appVersion, imageVersion).
		WithRegistryAuth(registry, username, password)

	tags := []string{appVersion, "latest"}
	addrs := []string{}
	for _, tag := range tags {
		addr, err := builtContainer.Publish(ctx, fmt.Sprintf("%s/watchedsky/server:%s", registry, tag))
		if err != nil {
			log.Println(err)
		}
		addrs = append(addrs, addr)
	}

	return addrs, nil
}
