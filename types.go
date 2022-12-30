package main

import "time"

type Image struct {
	Name  string    `json:"name"`
	Tag   string    `json:"tag"`
	Ctime time.Time `json:"ctime"`
}

type App struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type Flavor struct {
	Name    string   `json:"name"`
	Cpu     uint     `json:"cpu"`
	Mem     uint     `json:"memory"`
	Gpu     string   `json:"gpu,omitempty"` // OPTIONAL
	Regions []string `json:"regions,omitempty"`
}

type Region struct {
	Name string `json:"name"`
	Desc string `json:"desc,omitempty"` // OPTIONAL
}

type ConfigFile struct {
	Filename  string `json:"filename" yaml:"filename"`
	MountPath string `json:"mount_path" yaml:"mount_path"`
	Content   string `json:"content,omitempty" yaml:"content,omitempty"`
}

type ReleaseConfig struct {
	Name  string       `json:"name" yaml:"name"`
	Files []ConfigFile `json:"files" yaml:"files"`
}

type CreateReleaseArgs struct {
	Name     string   `json:"name" yaml:"name"`
	Desc     string   `json:"desc,omitempty" yaml:"desc,omitempty"` // OPTIONAL
	Image    string   `json:"image" yaml:"image"`
	Flavor   string   `json:"flavor" yaml:"flavor"`
	Port     uint     `json:"port,omitempty" yaml:"port,omitempty"` // OPRTIONAL
	Command  []string `json:"command,omitempty"`                    // OPTIONAL
	Args     []string `json:"args,omitempty"`                       // OPTIONAL
	HealthCk struct { // OPTIONAL
		Path    string `json:"path,omitempty" yaml:"path,omitempty"`
		Timeout uint   `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	} `json:"health_check,omitempty" yaml:"health_check,omitempty"`
	Env []struct {
		Key   string `json:"key" yaml:"key"`
		Value string `json:"value" yaml:"value"`
	} `json:"env,omitempty" yaml:"env,omitempty"` // OPTIONAL
	LogFilePaths []string       `json:"log_file_paths,omitempty" yaml:"log_file_paths,omitempty"` // OPTIONAL
	Config       *ReleaseConfig `json:"config,omitempty" yaml:"config,omitempty"`                 // OPTIONAL
}

type HealthCheck struct {
	Path    string `json:"path,omitempty"`
	Timeout uint   `json:"timeout,omitempty"` // OPTIONAL unit: second, default 3s
}

type EnvVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KodoInfo struct {
	BucketName string `json:"bucket_name"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
}

type Kodofs struct {
	Volume      string `json:"volume"`
	AccessToken string `json:"access_token"`
}

type Kodo struct {
	Goofys *KodoInfo `json:"goofys,omitempty"`
	Fcfs   *KodoInfo `json:"fcfs,omitempty"`
	Kodofs *Kodofs   `json:"kodofs,omitempty"`

	Region    string `json:"region"`
	MountPath string `json:"mount_path"`
	ReadOnly  bool   `json:"read_only,omitempty"`
}

type VolumeInfo struct {
	Kodo *Kodo `json:"kodo,omitempty"`
}

type Release struct {
	Name   string    `json:"name"`
	Desc   string    `json:"desc,omitempty"` // OPTIONAL
	Image  string    `json:"image"`
	Flavor string    `json:"flavor"`
	Port   uint      `json:"port,omitempty"`
	Ctime  time.Time `json:"ctime"` // RFC3339

	Command []string `json:"command,omitempty"` // OPTIONAL
	Args    []string `json:"args,omitempty"`    // OPTIONAL

	HealthCk     *HealthCheck   `json:"health_check,omitempty"` // OPTIONAL
	Env          []EnvVariable  `json:"env,omitempty"`
	Volumes      []VolumeInfo   `json:"volumes,omitempty"`        // OPTIONAL
	LogFilePaths []string       `json:"log_file_paths,omitempty"` // OPTIONAL
	Config       *ReleaseConfig `json:"config,omitempty"`
}

type Deploy struct {
	ID       string    `json:"id"`
	Release  string    `json:"release"`
	Region   string    `json:"region"`
	Replicas uint      `json:"replicas"`
	Ctime    time.Time `json:"ctime"`
}

type Instance struct {
	Ctime  time.Time `json:"ctime"`
	ID     string    `json:"id"`
	Status string    `json:"status"`
	IPs    string    `json:"ips,omitempty"`
}
