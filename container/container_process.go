package container

import (
	"os"
	"os/exec"
	"path"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
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
	// mntUrl := "/home/ourupf/mnt"
	// rootUrl := "/home/ourupf"
	// newWorkSpace(rootUrl, mntUrl)
	// cmd.Dir = mntUrl
	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, err
}

func newWorkSpace(rootUrl string, mntUrl string) {
	createReadOnlyLayer(rootUrl)
	createWriteLayer(rootUrl)
	createMntPoint(rootUrl, mntUrl)
}

func createReadOnlyLayer(rootUrl string) {
	busyboxUrl := path.Join(rootUrl, "busybox/")
	busyboxTarUrl := path.Join(rootUrl, "busybox.tar")
	exit, err := fileExist(busyboxUrl)
	if err != nil {
		log.Infof("fail to judge whether dir %s exist, err: %v", busyboxUrl, err)
	}
	if !exit {
		if err := os.Mkdir(busyboxUrl, 0777); err != nil {
			log.Errorf("make dir %s failed, error: %v", busyboxUrl, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarUrl, "-C", busyboxUrl).CombinedOutput(); err != nil {
			log.Errorf("untar failed, error: %v", err)
		}
	}
}

func createWriteLayer(rootUrl string) {
	writeUrl := path.Join(rootUrl, "writeLayer/")
	if err := os.Mkdir(writeUrl, 0777); err != nil {
		log.Errorf("create writelayer failed, error: %v", err)
	}
}

func createMntPoint(rootUrl string, mntUrl string) {
	log.Printf("create mount point")
	if err := os.Mkdir(mntUrl, 0777); err != nil {
		log.Errorf("mkdir dir %s failed, error: %v", mntUrl, err)
	}
	dirs := "dirs=" + path.Join(rootUrl, "writeLayer") + ":" + path.Join(rootUrl, "busybox")
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount aufs error: %v", err)
	}
}

func DeleteWorkSpace(rootUrl string, mntUrl string) {
	deleteMntPoint(mntUrl)
	deleteWriteLayer(rootUrl, mntUrl)
}

func deleteMntPoint(mntUrl string) {
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount error: %v", err)
	}
	cmd.Process.Wait()
}

func deleteWriteLayer(rootUrl string, mntUrl string) {
	if err := os.RemoveAll(mntUrl); err != nil {
		log.Errorf("delete mnt dir failed, error: %v", err)
	}
	writeUrl := path.Join(rootUrl, "writeLayer")
	if err := os.RemoveAll(writeUrl); err != nil {
		log.Errorf("remove write layer failed")
	}
}
