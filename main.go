package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"
)

type server struct {
}

// renderJSON renders 'v' as JSON and writes it as a response into w.
func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func NewServer() *server {
	return &server{}
}

func (s *server) pushImageHandler(w http.ResponseWriter, req *http.Request) {
	log.Print("handling push image\n")
	// Enforce a JSON Content-Type.
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	type RequestPushImage struct {
		Image string `json:"image"`
	}

	var img RequestPushImage
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&img); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if img.Image == "" {
		http.Error(w, "empty image", http.StatusBadRequest)
	}

	if exist, err := IsImageExists(img.Image); err == nil {
		if exist {
			return
		}
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = TryEnsureImageInDocker(img.Image); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = PushImage(img.Image); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *server) listImagesHandler(w http.ResponseWriter, req *http.Request) {
	images, err := ListImages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderJSON(w, images)
}

func IsImageExists(ref string) (bool, error) {
	if strings.ContainsAny(ref, "/") {
		parts := strings.Split(ref, "/")
		ref = parts[len(parts)-1]
	}
	images, err := ListImages()
	if err != nil {
		return false, err
	}
	for i := range images {
		if fmt.Sprintf("%s:%s", images[i].Name, images[i].Tag) == ref {
			return true, nil
		}
	}
	return false, nil
}

func (s *server) listAppsHandler(w http.ResponseWriter, req *http.Request) {
	apps, err := ListApps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderJSON(w, apps)
}

func (s *server) listFlavorsHandler(w http.ResponseWriter, req *http.Request) {
	flavors, err := ListFlavors()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderJSON(w, flavors)
}

func (s *server) listRegionsHandler(w http.ResponseWriter, req *http.Request) {
	regions, err := ListRegions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderJSON(w, regions)
}

func createRelease(app string, args *CreateReleaseArgs) error {
	ybytes, err := yaml.Marshal(args)
	if err != nil {
		return err
	}

	cfgDir, err := os.MkdirTemp("", "qapp-release-yml")
	if err != nil {
		return err
	}
	ymlfile, err := os.Create(filepath.Join(cfgDir, "dora.yaml"))
	if err != nil {
		return err
	}
	defer func() {
		ymlfile.Close()
		os.RemoveAll(cfgDir)
	}()
	if _, err = ymlfile.Write(ybytes); err != nil {
		return err
	}
	return QCreateRelease(app, cfgDir)
}

func (s *server) createReleaseHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling create release at %s\n", req.URL.Path)

	// Enforce a JSON Content-Type.
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	app := mux.Vars(req)["app"]

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var args CreateReleaseArgs
	if err := dec.Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if args.Name == "" {
		args.Name = fmt.Sprintf("%s%s", "v", time.Now().Format("060102-150405"))
	}

	if args.Image == "" {
		http.Error(w, "empty image", http.StatusBadRequest)
		return
	} else {
		exist, err := IsImageExists(args.Image)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !exist {
			http.Error(w, fmt.Sprintf("image '%s' not exists, please upload firstly", args.Image), http.StatusBadRequest)
			return
		}
	}

	type ResponseName struct {
		Name string `json:"name"`
	}

	if err = createRelease(app, &args); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderJSON(w, ResponseName{Name: args.Name})
}

func (s *server) listReleasesHandler(w http.ResponseWriter, req *http.Request) {
	log.Print("handling list releases\n")
	app := mux.Vars(req)["app"]

	releases, err := ListReleases(app)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderJSON(w, releases)
}

func (s *server) listDeploysHandler(w http.ResponseWriter, req *http.Request) {
	log.Print("handling list deploys\n")
	app := mux.Vars(req)["app"]
	region := mux.Vars(req)["region"]
	releaseValues := req.URL.Query()["release"]

	if region == "" {
		http.Error(w, "empty region", http.StatusBadRequest)
		return
	}

	deploys, err := ListDeploys(app, region)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var ret = make([]Deploy, 0)
	if len(releaseValues) == 0 {
		ret = deploys
	} else {
		for i := range deploys {
			if deploys[i].Release == releaseValues[0] {
				ret = append(ret, deploys[i])
			}
		}
	}
	renderJSON(w, ret)
}

func (s *server) createDeployHandler(w http.ResponseWriter, req *http.Request) {
	log.Print("handling create deploy\n")

	// Enforce a JSON Content-Type.
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	app := mux.Vars(req)["app"]

	type RequestDeploy struct {
		Release  string `json:"release"`
		Region   string `json:"region"`
		Replicas uint   `json:"replicas"`
	}
	type ResponseDeploy struct {
		ID string `json:"id"`
	}

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var args RequestDeploy
	if err := dec.Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if args.Region == "" || args.Release == "" {
		http.Error(w, "empty region or release", http.StatusBadRequest)
		return
	}

	deploy, err := CreateDeploy(app, args.Release, args.Region, int(args.Replicas))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderJSON(w, ResponseDeploy{ID: deploy.ID})
}

func (s *server) deleteDeployHandler(w http.ResponseWriter, req *http.Request) {
	log.Print("handling delete deploy\n")

	// Enforce a JSON Content-Type.
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	app := mux.Vars(req)["app"]

	type RequestDeploy struct {
		ID     string `json:"id"`
		Region string `json:"region"`
	}

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var args RequestDeploy
	if err := dec.Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if args.Region == "" || args.ID == "" {
		http.Error(w, "empty region or ID", http.StatusBadRequest)
		return
	}

	err = DeleteDeploy(app, args.ID, args.Region)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *server) listDeployInstancesHandler(w http.ResponseWriter, req *http.Request) {
	log.Print("handling list instances\n")
	app := mux.Vars(req)["app"]
	deployID := mux.Vars(req)["deploy"]
	region := mux.Vars(req)["region"]

	if region == "" {
		http.Error(w, "empty region", http.StatusBadRequest)
		return
	}

	instances, err := ListInstance(app, deployID, region)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderJSON(w, instances)
}

func main() {
	var ak string
	var sk string
	var addr string

	flag.StringVar(&ak, "access-key", "", "access key of Qiniu account")
	flag.StringVar(&sk, "secret-key", "", "secret key of Qiniu account")
	flag.StringVar(&addr, "listen-addr", ":9100", "HTTP listen adress, i.e 0.0.0.0:9100")
	flag.Parse()
	if ak == "" || sk == "" {
		if ak == "" {
			ak, _ = os.LookupEnv("ACCESS_KEY")
		}
		if sk == "" {
			sk, _ = os.LookupEnv("SECRET_KEY")
		}
	}
	err := Login(ak, sk)
	if err != nil {
		log.Fatalf("login failed, please check your access-key/secret-key pair: %v", err)
	}

	router := mux.NewRouter()
	router.StrictSlash(true)
	server := NewServer()

	router.HandleFunc("/images", server.listImagesHandler).Methods("GET")
	router.HandleFunc("/images", server.pushImageHandler).Methods("POST")

	router.HandleFunc("/apps", server.listAppsHandler).Methods("GET")
	router.HandleFunc("/flavors", server.listFlavorsHandler).Methods("GET")
	router.HandleFunc("/regions", server.listRegionsHandler).Methods("GET")

	router.HandleFunc("/apps/{app:[a-z](?:[-a-z0-9]*[a-z0-9])}/releases", server.listReleasesHandler).Methods("GET")
	router.HandleFunc("/apps/{app:[a-z](?:[-a-z0-9]*[a-z0-9])}/releases", server.createReleaseHandler).Methods("POST")

	router.HandleFunc("/apps/{app:[a-z](?:[-a-z0-9]*[a-z0-9])}/deploys", server.listDeploysHandler).Methods("GET").Queries("region", "{region}")
	router.HandleFunc("/apps/{app:[a-z](?:[-a-z0-9]*[a-z0-9])}/deploys", server.createDeployHandler).Methods("POST")
	router.HandleFunc("/apps/{app:[a-z](?:[-a-z0-9]*[a-z0-9])}/deploys", server.deleteDeployHandler).Methods("DELETE")

	router.HandleFunc("/apps/{app:[a-z](?:[-a-z0-9]*[a-z0-9])}/deploys/{deploy}/instances", server.listDeployInstancesHandler).Queries("region", "{region}").Methods("GET")

	log.Fatal(http.ListenAndServe(addr, router))
}
