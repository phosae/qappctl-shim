package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

func run(cmd *exec.Cmd, action string) ([]byte, error) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("err %s: %s, %s", action, err, string(out))
	}
	return out, nil
}

func Login(ak, sk string) error {
	cmd := exec.Command("qappctl", "login", "--ak", ak, "--sk", sk)
	_, err := run(cmd, "login")
	if err != nil {
		return err
	}
	return nil
}

func PushImage(ref string) error {
	cmd := exec.Command("qappctl", "push", ref)
	_, err := run(cmd, "push image "+ref)
	return err
}

// qappctl images -o json
func ListImages() (images []Image, err error) {
	cmd := exec.Command("qappctl", "images", "-o", "json")
	out, err := run(cmd, "list images")
	if err != nil {
		return nil, err
	}
	return images, json.Unmarshal(out, &images)
}

// qappctl release list <app> -o json
func ListReleases(app string) (releases []Release, err error) {
	cmd := exec.Command("qappctl", "release", "list", app, "-o", "json")
	out, err := run(cmd, "list images")
	if err != nil {
		return nil, err
	}
	return releases, json.Unmarshal(out, &releases)
}

// qappctl release create <app> -c <config-dir>
func QCreateRelease(app, cfgdir string) (err error) {
	cmd := exec.Command("qappctl", "release", "create", app, "-c", cfgdir)
	_, err = run(cmd, "create release")
	return err
}

// qappctl deploy list --region <region> <app>
func ListDeploys(app, region string) (deploys []Deploy, err error) {
	cmd := exec.Command("qappctl", "deploy", "list", "--region", region, app, "-o", "json")
	out, err := run(cmd, "list deploys")
	if err != nil {
		return nil, err
	}
	return deploys, json.Unmarshal(out, &deploys)
}

// qappctl deploy create <app> --release <release> --region <region> --expect_replicas <num>
//
//	{
//	  "id": "h221201-1658-30080-p8gv",
//	  "release": "zenx-v0",
//	  "region": "z0",
//	  "replicas": 0,
//	  "ctime": "0001-01-01T00:00:00Z"
//	}
func CreateDeploy(app, release, region string, replicas int) (deploy *Deploy, err error) {
	cmd := exec.Command("qappctl", "deploy", "create", app,
		"--region", region, "--release", release,
		"--expect_replicas", strconv.Itoa(replicas),
		"-o", "json")
	out, err := run(cmd, "create deploy")
	if err != nil {
		return nil, err
	}
	return deploy, json.Unmarshal(out, &deploy)
}

// qappctl deploy delete <app> --id <deploy_id> --region <region>
func DeleteDeploy(app, id, region string) error {
	cmd := exec.Command("qappctl", "deploy", "delete", app, "--id", id, "--region", region)
	if _, err := run(cmd, "create deploy"); err != nil {
		return err
	}
	return nil
}

// qappctl instance list <app> --deploy <deployID> --region <region>
func ListInstance(app, id, region string) (ins []Instance, err error) {
	cmd := exec.Command("qappctl", "instance", "list", app, "--deploy", id, "--region", region, "-o", "json")
	out, err := run(cmd, "list deploy instances")
	if err != nil {
		return nil, err
	}
	return ins, json.Unmarshal(out, &ins)
}
