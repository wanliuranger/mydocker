package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func RunContainerInitProcess() error {

	mntUrl := "/home/ourupf/mnt"
	rootUrl := "/home/ourupf"
	newWorkSpace(rootUrl, mntUrl)
	os.Chdir(mntUrl)

	commandArr := readUserCommand()
	if len(commandArr) == 0 {
		return fmt.Errorf("get user command error, nothing in cmdArr")
	}

	// defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	setupMount()

	path, err := exec.LookPath(commandArr[0])
	if err != nil {
		log.Errorf("exec look path error: %v", err)
		return err
	}

	log.Info("Find path: %s", path)

	if err := syscall.Exec(path, commandArr, os.Environ()); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func pivot_root(root string) error {
	// if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
	// 	return fmt.Errorf("mount rootfs to itself error: %v", err)
	// }
	pivotDir := path.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return fmt.Errorf("create pivotDir error: %v", err)
	}
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("new mount space error: %v", err)
	}
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot root error: %v", err)
	}
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir error: %v", err)
	}
	pivotDir = path.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmout error: %v", err)
	}
	return os.RemoveAll(pivotDir)
}

func setupMount() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("can't get current location: %v", err)
		return
	}
	log.Infof("current location is: %v", pwd)
	err = pivot_root(pwd)
	if err != nil {
		log.Printf("pivot error: %v", err)
	}

	// defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")

}
