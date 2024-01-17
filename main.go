package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/directory"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/types"
)

type config struct {
	Locations []location
}

type OciManifest struct {
	SchemaVersion int             `json:"schemaVersion"`
	MediaType     string          `json:"mediaType"`
	Config        mediaMetadata   `json:"config"`
	Layers        []mediaMetadata `json:"layers"`
}

type mediaMetadata struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
}

type location struct {
	Name        string `json:"name"`
	Dst string `json:"dst"`
	Src string `json:"src"`
}

func main() {
	homedir := os.Getenv("HOME")
	workdir := fmt.Sprintf("%s/.local/share/ImmutableDotfileManager", homedir)
	filedata, err := os.ReadFile(fmt.Sprintf("%s/.idotfiles.json", homedir))
	check(err)
	conf := config{}
	json.Unmarshal([]byte(filedata), &conf)

	if _, err := os.Stat(workdir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll((workdir), 0600)
		}
	}


	for _, i := range conf.Locations {
		manifest := getImage(workdir, i.Name)
		var lowerdir string
		for _, v := range manifest.Layers {
			dir := fmt.Sprintf("%s/filesystems/%s/%s", workdir, i.Name, extractDigest(v.Digest))
			lowerdir = fmt.Sprintf("%s,%s", lowerdir, dir)
		}
		overlayWorkdir := fmt.Sprintf("%s/filesystems/%s/workdir", workdir, i.Name)
		mutable := fmt.Sprintf("%s/filesystems/%s/mutbale", workdir, i.Name)

		finalCommand := fmt.Sprintf("-t overlay overlay -o lowerdir=%s,workdir=%s,upperdir=%s", lowerdir, overlayWorkdir, mutable)
		exec := exec.Command("mount", finalCommand)
		exec.Wait()
		output, err := exec.Output()
		check(err)

		fmt.Println(string(output))
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func getImage(dir string, name string) OciManifest {
	if _, err := os.Stat(fmt.Sprintf("%s/images/%s", dir, name)); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll((fmt.Sprintf("%s/images/%s", dir, name)), 0600)
		}
	}
	ctx := context.Background()
	polctx, err := signature.NewPolicyContext(&signature.Policy{
		Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()},
	})
	check(err)
	src, err := docker.ParseReference("//ubuntu:latest")
	check(err)
	dst, err := directory.Transport.ParseReference(fmt.Sprintf("//%s/images/%s", dir, name))
	check(err)

	dstctx := types.SystemContext{
		DirForceDecompress: true,
	}

	options := copy.Options{
		DestinationCtx: &dstctx,
	}

	manifest, err := copy.Image(ctx, polctx, dst, src, &options)
	check(err)

	string := string(manifest[:])
	fmt.Println(string)

	var out OciManifest
	errora := json.Unmarshal(manifest, &out)
	check(errora)

	for _, v := range out.Layers {
		Extract(v.Digest, dir, name)
	}

	return out
}

func Extract(layerdigest string, dir string, name string) {
	if _, err := os.Stat(fmt.Sprintf("%s/filsystems/%s", dir, name)); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll((fmt.Sprintf("%s/filsystems/%s", dir, name)), 0600)
		}
	}

	layerdigestnew := strings.Replace(layerdigest, "sha256:", "", 1)
	if _, err := os.Stat(fmt.Sprintf("%s/filsystems/%s", dir, name)); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll((fmt.Sprintf("%s/filsystems/%s", dir, name)+"/"+layerdigestnew), 0600)
		}
	} else {
		os.Remove(fmt.Sprintf("%s/filsystems/%s", dir, name)+"/"+layerdigestnew)
		os.MkdirAll((fmt.Sprintf("%s/filsystems/%s", dir, name)+"/"+layerdigestnew), 0600)
	}

	untarCmd := exec.Command("tar", "-xf", fmt.Sprintf("%s/images/%s/%s",dir, name, layerdigestnew), "-C", fmt.Sprintf("%s/filesystems/%s/%s",dir,name,layerdigestnew))

	untarCmd.Output()
}

func extractDigest(fulldigest string) string {
	out := strings.Replace(fulldigest, "sha256:", "", 1)
	return out
}