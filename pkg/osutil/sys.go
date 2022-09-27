package osutil

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/llaoj/kube-finder/internal/config"
	log "github.com/sirupsen/logrus"
)

func ContainerPID(containerID string) string {
	// /sys/fs/cgroup/pids...<containerID>.../cgroup.procs
	pid := ""
	_ = filepath.Walk(config.Get().CgroupPIDRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if strings.Contains(path, containerID) && strings.Contains(path, "cgroup.procs") {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				log.Error(err)
				return err
			}
			procs := string(content)
			log.WithFields(log.Fields{
				"path":         path,
				"cgroup.procs": procs,
			}).Trace()
			i := strings.Index(procs, "\n")
			if i == -1 {
				pid = procs
			} else {
				pid = procs[:i]
			}
			return errors.New("ok")
		}
		return nil
	})

	return pid
}
