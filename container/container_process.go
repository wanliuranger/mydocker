package container

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const (
	rootUrl        = "/home/ourupf"
	mntBase        = "/home/ourupf/mnt"
	writeLayerBase = "/home/ourupf/writeLayer"
)

func NewParentProcess(tty bool, containerName string, imageName string, volume string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("new pipe error")
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	newWorkSpace(containerName, imageName, volume)
	cmd.Dir = path.Join(mntBase, containerName)
	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, err
}

func newWorkSpace(containerName string, imageName string, volume string) {
	createReadOnlyLayer(imageName)
	createWriteLayer(containerName)
	createMntPoint(containerName, imageName)
	mountVolume(containerName, volume)
}

func mountVolume(containerName string, volume string) {
	if volume != "" {
		urls := strings.Split(volume, ":")
		if len(urls) == 2 && urls[0] != "" && urls[1] != "" {
			exist, err := fileExist(urls[0])
			if err != nil {
				log.Errorf("fail to judge where dir %s exist, err:%v", urls[0], err)
			}
			if !exist {
				if err := os.Mkdir(urls[0], 0777); err != nil {
					log.Errorf("fail to create volume %s, err: %v", urls[0], err)
				}
			}

			mntPath := path.Join(mntBase, containerName, urls[1])
			if err := os.Mkdir(mntPath, 0777); err != nil {
				log.Errorf("fail to create mount point %s, err: %v", mntPath, err)
			}

			dirs := "dirs=" + urls[0]
			cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntPath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Errorf("mount volume error: %v", err)
			}

		} else {
			log.Errorf("wrong volume format")
		}
	}
}

func createReadOnlyLayer(imageName string) {
	imageUrl := path.Join(rootUrl, imageName)
	imageTarUrl := path.Join(rootUrl, imageName+".tar")
	exit, err := fileExist(imageUrl)
	if err != nil {
		log.Errorf("fail to judge whether dir %s exist, err: %v", imageUrl, err)
	}
	if !exit {
		if err := os.Mkdir(imageUrl, 0777); err != nil {
			log.Errorf("make dir %s failed, error: %v", imageUrl, err)
		}
		if _, err := exec.Command("tar", "-xvf", imageTarUrl, "-C", imageUrl).CombinedOutput(); err != nil {
			log.Errorf("untar failed, error: %v", err)
		}
	}
}

func createWriteLayer(containerName string) {
	writeUrl := path.Join(writeLayerBase, containerName)
	if err := os.Mkdir(writeUrl, 0777); err != nil {
		log.Errorf("create writelayer failed, error: %v", err)
	}
}

func createMntPoint(containerName string, imageName string) {
	log.Printf("create mount point")
	mntUrl := path.Join(mntBase, containerName)
	if err := os.Mkdir(mntUrl, 0777); err != nil {
		log.Errorf("mkdir dir %s failed, error: %v", mntUrl, err)
	}
	dirs := "dirs=" + path.Join(writeLayerBase, containerName) + ":" + path.Join(rootUrl, imageName)
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount aufs error: %v", err)
	}
}

func DeleteWorkSpace(containerName string, volume string) {
	umountVolume(containerName, volume)
	deleteMntPoint(containerName)
	deleteWriteLayer(containerName)
}

func deleteMntPoint(containerName string) {
	mntUrl := path.Join(mntBase, containerName)
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount error: %v", err)
	}
	cmd.Process.Wait()
}

func deleteWriteLayer(containerName string) {
	mntUrl := path.Join(mntBase, containerName)
	if err := os.RemoveAll(mntUrl); err != nil {
		log.Errorf("delete mnt dir failed, error: %v", err)
	}
	writeUrl := path.Join(writeLayerBase, containerName)
	if err := os.RemoveAll(writeUrl); err != nil {
		log.Errorf("remove write layer failed")
	}
}

func umountVolume(containerName string, volume string) {
	if volume != "" {
		urls := strings.Split(volume, ":")
		if len(urls) == 2 && urls[0] != "" && urls[1] != "" {
			volumePath := path.Join(mntBase, containerName, urls[1])
			cmd := exec.Command("umount", volumePath)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				log.Errorf("umount volume %s failed, err: %v", urls[1], err)
			}
			cmd.Process.Wait()
		}
	}
}
