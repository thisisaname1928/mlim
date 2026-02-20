package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func mountSystemfs(rootPath string) {
	e := syscall.Mount("proc", rootPath+"/"+"proc", "proc", uintptr(0), "")
	if e != nil {
		fmt.Println("Mounting procfs failed!")
	}
	e = syscall.Mount("sysfs", rootPath+"/"+"sys", "sysfs", uintptr(0), "")
	if e != nil {
		fmt.Println("Mounting sysfs failed!")
	}
	e = syscall.Mount("devtmpfs", rootPath+"/"+"dev", "devtmpfs", uintptr(0), "")
	if e != nil {
		fmt.Println("Mounting devtmpfs failed!")
	}
	e = syscall.Mount("tmpfs", rootPath+"/"+"tmp", "tmpfs", uintptr(0), "")
	if e != nil {
		fmt.Println("Mounting tmpfs failed!")
	}
}

func umountSystemfs(rootPath string) {
	e := syscall.Unmount(rootPath+"/"+"proc", 0)
	if e != nil {
		fmt.Println("Unmount proc failed!")
	}
	e = syscall.Unmount(rootPath+"/"+"sys", 0)
	if e != nil {
		fmt.Println("Unmount sys failed!")
	}
	e = syscall.Unmount(rootPath+"/"+"dev", 0)
	if e != nil {
		fmt.Println("Unmount dev failed!")
	}
	e = syscall.Unmount(rootPath+"/"+"tmp", 0)
	if e != nil {
		fmt.Println("Unmount tmp failed!")
	}
}

func main() {
	fmt.Print("\033[2J\033[H")
	fmt.Println("hello world!")

	mountSystemfs("")

	e := syscall.Mount("/dev/sda", "/newRoot", "ext4", 0, "")
	umountSystemfs("")
	mountSystemfs("/newRoot")

	if e != nil {
		panic(e)
	}

	// free initramfs
	dir, e := os.ReadDir("/")
	for _, v := range dir {
		if v.Name() == "newRoot" || v.Name() == "init" {
			continue
		}

		e := os.RemoveAll("/" + v.Name())
		if e != nil {
			fmt.Println("Remove", v.Name(), "got:", e)
		}

	}

	os.Chdir("/newRoot")
	e = syscall.Chroot("/newRoot")
	os.Chdir("/")
	fmt.Println(e)

	dir, e = os.ReadDir("/")
	for _, v := range dir {
		fmt.Print(v.Name(), " ")
	}

	fmt.Println("switch root ok!")

	console, e := os.OpenFile("/dev/console", os.O_RDWR, 0)

	if e != nil {
		fmt.Println(e)
	}
	defer console.Close()

	cmd := exec.Command("/sbin/umag")
	cmd.Stdout = console
	cmd.Stdin = console
	cmd.Stderr = console
	e = cmd.Run()
	if e != nil {
		fmt.Println(e)
	}

	for {
	}
}
