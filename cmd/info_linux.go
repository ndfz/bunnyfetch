//go:build linux
// +build linux

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func OS() string {
	return os.Getenv("NAME")
}

func Kernel() string {
	// /proc/version should always exist on linux
	procver, _ := os.ReadFile("/proc/version")

	// /proc/version has the same format with "Linux version <kern-version>" as the 3rd
	// word, `procver` is []byte so has to be converted
	return strings.Split(string(procver), " ")[2]
}

func Shell() string {
	shellenv := strings.Split(os.Getenv("SHELL"), "/")
	return shellenv[len(shellenv)-1]
}

func WM() string {
	if waylandDisplay := os.Getenv("WAYLAND_DISPLAY"); os.Getenv("XDG_SESSION_TYPE") == "wayland" || waylandDisplay != "" {
		sockPath := os.Getenv("XDG_RUNTIME_DIR") + "/" + waylandDisplay
		_, err := os.Stat(sockPath)
		if err != nil {
			return "Unknown"
		}

		// TODO: not shell out to fuser
		// doing its job is half hell and i dont really want to deal with it
		out, err := exec.Command("fuser", sockPath).CombinedOutput()
		if err != nil {
			return "Unknown"
		}

		procID := strings.TrimSpace(strings.Split(string(out), ":")[1])
		bin, err := os.Readlink("/proc/" + procID + "/exe")
		// apparently may not exist?
		if err != nil {
			return "Unknown"
		}

		splitBin := strings.Split(bin, "/")
		return splitBin[len(splitBin)-1]
	}
	X, err := xgb.NewConn()
	if err != nil {
		return "Unknown"
	}
	defer X.Close()

	setup := xproto.Setup(X)
	root := setup.DefaultScreen(X).Root

	// get a "window" for the window manager (makes sense huh)
	aname := "_NET_SUPPORTING_WM_CHECK"
	activeAtom, err := xproto.InternAtom(X, true, uint16(len(aname)), aname).Reply()
	if err != nil {
		return "Unknown"
	}

	// get the atom for the actual window manager name
	aname = "_NET_WM_NAME"
	nameAtom, err := xproto.InternAtom(X, true, uint16(len(aname)), aname).Reply()
	if err != nil {
		return "Unknown"
	}

	// and its value
	reply, err := xproto.GetProperty(X, false, root, activeAtom.Atom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return "Unknown"
	}
	windowID := xproto.Window(xgb.Get32(reply.Value))

	reply, err = xproto.GetProperty(X, false, windowID, nameAtom.Atom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return "Unknown"
	}

	return string(reply.Value)
}

func Terminal() string {
	var term string

	// Check parent processes
	parent := os.Getppid()
	for term == "" {
		ppid := fmt.Sprintf("%d", parent)
		ppid, err := getPPID(ppid)
		if err != nil || ppid == "" {
			break
		}
		parentName, err := getProcessName(ppid)
		if err != nil || parentName == "" {
			break
		}

		switch parentName {
		case "sh", "bash", "zsh", "screen", "su", "newgrp":
		case "login", "init", "systemd", "sshd":
			term = "tty"
		case "gnome-terminal-":
			term = "gnome-terminal"
		case "urxvtd":
			term = "urxvt"
		case "nvim":
			term = "Neovim Terminal"
		case "NeoVimServer":
			term = "VimR Terminal"
		default:
			if filepath.Base(parentName) == parentName {
				term = parentName
			}
			if strings.HasSuffix(term, "-wrapped") {
				term = strings.TrimSuffix(term, "-wrapped")
			}
		}

		if term != "" {
			break
		}
		parent, _ = strconv.Atoi(ppid)
	}

	return term
}

// getPPID returns the parent process ID of a given PID.
func getPPID(pid string) (string, error) {
	cmd := exec.Command("ps", "-p", pid, "-o", "ppid=")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getProcessName returns the name of the process given its PID.
func getProcessName(pid string) (string, error) {
	cmd := exec.Command("ps", "-p", pid, "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
