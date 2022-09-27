package finder

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/llaoj/kube-finder/pkg/kube"
	"github.com/llaoj/kube-finder/pkg/osutil"
	log "github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type Controller struct {
	EndpointManager *EndpointManager
}

func NewController() *Controller {
	namespace := os.Getenv("KUBE_NAMESPACE")
	service := os.Getenv("KUBE_SERVICE")
	labelSelector, err := kube.ServiceLabelSelector(context.Background(), namespace, service)
	log.WithFields(log.Fields{"labelSelector": labelSelector}).Trace()
	if err != nil {
		log.Fatal(err)
	}
	endpointManager := NewEndpointManager(namespace, labelSelector)
	controller := &Controller{
		EndpointManager: endpointManager,
	}

	return controller
}

func (controller *Controller) ProxyHandler(c *gin.Context) {
	namespaceName := c.Param("namespace")
	podName := c.Param("pod")
	containerName := c.Param("container")

	pod, err := kube.Pod(c, namespaceName, podName)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if pod == nil {
		c.String(http.StatusNotFound, "pod not found")
		return
	}

	containerID := ""
	for _, item := range pod.Status.ContainerStatuses {
		if item.Name == containerName {
			containerID = kube.ParseContainerID(item.ContainerID)
			break
		}
	}
	if containerID == "" {
		c.String(http.StatusNotFound, "containerid not found")
		return
	}
	log.WithFields(log.Fields{
		"containerID": containerID,
		"hostIP":      pod.Status.HostIP,
	}).Trace()

	// reverse proxy
	log.WithFields(log.Fields{"endpoints": controller.EndpointManager.Endpoints}).Tracef("kube-finder endpoints")
	endpoint, exist := controller.EndpointManager.Endpoints[pod.Status.HostIP]
	if !exist {
		log.WithFields(log.Fields{"endpoint": endpoint}).Error("finder endpoint not found")
		c.String(http.StatusNotFound, "finder endpoint not found")
		return
	}
	url := "http://" + endpoint + "/apis/v1/containers/" + containerID + "/files?" + c.Request.URL.RawQuery
	log.Debugf("reverse proxy url: %v", url)
	proxyEndpoint(c, url)
	return
}

func proxyEndpoint(c *gin.Context, rawURL string) {
	client := &http.Client{}
	req, err := http.NewRequest(c.Request.Method, rawURL, c.Request.Body)
	if err != nil {
		c.String(http.StatusServiceUnavailable, "new request failed: %s", err)
		return
	}
	req.Header = c.Request.Header
	log.WithFields(log.Fields{"request": req}).Debug("reverse proxy...")
	resp, err := client.Do(req)
	if err != nil {
		c.String(http.StatusServiceUnavailable, "reverse proxy failed: %s", err)
		return
	}
	reader := resp.Body
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)
	contentLength := resp.ContentLength
	contentType := resp.Header.Get("Content-Type")
	c.DataFromReader(resp.StatusCode, contentLength, contentType, reader, nil)
	return
}

func (controller *Controller) ListHandler(c *gin.Context) {
	containerID := c.Param("containerid")
	subpath := clipSubpath(c.DefaultQuery("subpath", "/"))

	pid := osutil.ContainerPID(containerID)
	if pid == "" {
		c.String(http.StatusNotFound, "pid not found")
		return
	}

	fullPath := fmt.Sprintf("/host/proc/%s/root%s", pid, subpath)
	log.WithFields(log.Fields{"fullPath": fullPath}).Trace()
	if !osutil.Exists(fullPath) {
		c.String(http.StatusNotFound, "subpath not found")
		return
	}

	if osutil.IsDir(fullPath) {
		files, err := ioutil.ReadDir(fullPath)
		if err != nil {
			log.Error(err)
			c.Status(http.StatusInternalServerError)
			return
		}

		var responseFiles ResponseFiles
		responseFiles.Subpath = subpath
		for _, file := range files {
			f, err := statFile(fullPath, file)
			if err != nil {
				c.String(http.StatusServiceUnavailable, err.Error())
				return
			}
			responseFiles.Files = append(responseFiles.Files, *f)
		}
		c.JSON(http.StatusOK, responseFiles)
		return
	} else {
		c.File(fullPath)
		return
	}
}

func clipSubpath(subpath string) string {
	if !strings.HasPrefix(subpath, "/") {
		subpath = "/" + subpath
	}
	if len(subpath) > 1 {
		subpath = strings.TrimRight(subpath, "/")
	}
	return subpath
}

func statFile(path string, f fs.FileInfo) (*File, error) {
	file := &File{
		Name:    f.Name(),
		Size:    f.Size(),
		Mode:    f.Mode().String(),
		ModTime: f.ModTime(),
		IsDir:   f.IsDir(),
	}

	stat, ok := f.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, errors.New("cannot convert Sys() value to syscall.Stat_t")
	}

	root := strings.Join(strings.Split(path, "/")[:5], "/")
	u, err := osutil.LookupUserIdFrom(root+"/etc/passwd", fmt.Sprint(stat.Uid))
	if err != nil {
		return nil, err
	}
	file.UserName = u.Username

	group, err := osutil.LookupGroupIdFrom(root+"/etc/group", fmt.Sprint(stat.Gid))
	if err != nil {
		return nil, err
	}
	file.GroupName = group.Name

	if f.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(fmt.Sprintf("%s/%s", path, f.Name()))
		if err != nil {
			return nil, err
		}
		file.Link = link
	}

	return file, nil
}

func (controller *Controller) CreateHandler(c *gin.Context) {
	containerID := c.Param("containerid")
	subpath := clipSubpath(c.DefaultQuery("subpath", "/"))
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "get file err: %s", err.Error())
		return
	}

	pid := osutil.ContainerPID(containerID)
	if pid == "" {
		c.String(http.StatusNotFound, "pid not found")
		return
	}

	filename := fmt.Sprintf("/host/proc/%s/root%s/%s", pid, subpath, filepath.Base(file.Filename))
	log.WithFields(log.Fields{"filename": filename}).Debug("creating file...")
	// 0666
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
		return
	}
	// 0777
	if err := os.Chmod(filename, os.FileMode(0777)); err != nil {
		c.String(http.StatusBadRequest, "file chmod err: %s", err.Error())
		return
	}

	c.String(http.StatusOK, "file %s uploaded successfully", filename)
}
