// SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/gin-gonic/gin"
)

var programLevel = new(slog.LevelVar)
var appVersion = "v0.0.0-dev"
var defaultSessionKey = "REPLACE-ME-*H)dC/),{%;6&zrr(almasdr3SFAE2"

const inboundFolder = "./var/inbound/"

//go:embed assets/* templates/*
var f embed.FS

func main() {
	var err error
	var address string
	var verbosity, mode, sessionKey, socketPath string
	var fd int
	flag.StringVar(&address, "address", "127.0.0.1:8080", "port to listen on")
	flag.StringVar(&verbosity, "verbosity", "info", "verbosity level")
	flag.StringVar(&mode, "mode", "release", "mode to run in")
	flag.StringVar(&sessionKey, "session", defaultSessionKey, "session key (SESSION_KEY)")
	flag.StringVar(&socketPath, "socket", "", "unix socket path")

	flag.Parse()

	if !isFlagPassed("session") {
		e := os.Getenv("SESSION_KEY")
		if e != "" {
			sessionKey = e
		}
	}

	setupLog(verbosity)

	slog.Info("starting unionfees-server", "version", appVersion, "mode", mode, "verbosity", verbosity)
	fd, err = getSystemdSocketHandle()
	if err != nil {
		slog.Info("not running under systemd, using default address")
		fd = 0 // no file descriptor available
	}

	switch mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	templ := template.Must(
		template.New("").ParseFS(f,
			"templates/*.tmpl",
		// "templates/foo/*.tmpl"
		))

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8 MiB
	r.SetHTMLTemplate(templ)
	err = setupSessions(r, sessionKey)
	if err != nil {
		slog.Error("error setting up sessions", "error", err)
		os.Exit(1)
	}

	r.StaticFS("/public", http.FS(f))

	r.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		files := session.Get("files")
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":   "Unionfees Server",
			"version": appVersion,
			"flashes": session.Flashes(),
			"files":   files,
		})
	})

	r.GET("/ping/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/upload/", parseToFirstTxtHandler)
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/public/assets/favicon.ico")
	})

	if fd > 0 {
		slog.Info("starting server on file descriptor", "fd", fd)
		// use the file descriptor to listen on
		if err := r.RunFd(fd); err != nil {
			slog.Error("error starting server on file descriptor", "fd", fd, "error", err)
			os.Exit(1)
		}
	} else if socketPath != "" {
		slog.Info("starting server", "socketPath", socketPath)
		// always remove the named socket from the fs if its there
		err = syscall.Unlink(socketPath)
		if err != nil {
			// not really important if it fails
			slog.Error("Unlink()", "error", err)
		}
		listener, err := net.Listen("unix", socketPath)
		if err != nil {
			slog.Error("error creating unix socket listener", "socketPath", socketPath, "error", err)
			os.Exit(1)
		}
		defer listener.Close()
		defer os.Remove(socketPath)
		os.Chmod(socketPath, 0660) // set permissions to allow access

		if err := r.RunUnix(socketPath); err != nil {
			slog.Error("error starting server on unix socket", "error", err)
			os.Exit(1)
		}

	} else {
		slog.Info("starting server", "address", address, "url", fmt.Sprintf("http://%s", address))

		// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
		if err := r.Run(address); err != nil {
			slog.Error("error starting server", "error", err)
			os.Exit(1)
		}
	}
}

func setupLog(verbosity string) {
	switch verbosity {
	case "debug":
		programLevel.Set(slog.LevelDebug)
	case "info":
		programLevel.Set(slog.LevelInfo)
	case "warn":
		programLevel.Set(slog.LevelWarn)
	case "error":
		programLevel.Set(slog.LevelError)
	default:
		programLevel.Set(slog.LevelInfo)
	}
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(handler))
}

func setupSessions(r *gin.Engine, sessionKey string) error {
	if sessionKey == "" {
		return fmt.Errorf("session key is empty")
	}
	slog.Debug("setting up sessions")
	store := cookie.NewStore([]byte(sessionKey))

	r.Use(sessions.Sessions("unionfees", store))
	return nil
}

func getSystemdSocketHandle() (int, error) {
	// https://www.freedesktop.org/software/systemd/man/sd_listen_fds.html
	const SdListenFdsStart = 3
	ListenPID := os.Getenv("LISTEN_PID")
	if ListenPID == "" || ListenPID != fmt.Sprintf("%d", os.Getpid()) {
		return 0, fmt.Errorf("not running under systemd")
	}
	ListenFDS := os.Getenv("LISTEN_FDS")
	if ListenFDS == "" {
		return 0, fmt.Errorf("LISTEN_FDS environment variable is not set")
	}
	fds, err := strconv.Atoi(ListenFDS)
	if err != nil {
		return 0, fmt.Errorf("invalid LISTEN_FDS value: %v", err)
	}
	fdnames := os.Getenv("LISTEN_FDNAMES")
	slog.Info("found LISTEN_FDS environment variable",
		"LISTEN_FDS", ListenFDS,
		"LISTEN_PID", ListenPID,
		"LISTEN_FDNAMES", fdnames)

	isSocketActivated := ListenPID != "" && ListenFDS != ""

	if !isSocketActivated {
		return 0, fmt.Errorf("not socket activated, LISTEN_PID or LISTEN_FDS is not set")
	}

	lpid, err := strconv.Atoi(ListenPID)
	if err != nil {
		return 0, fmt.Errorf("invalid LISTEN_PID value: %v", err)
	}
	if lpid != os.Getpid() {
		return 0, fmt.Errorf("LISTEN_PID does not match current process ID, expected %d, got %d", os.Getpid(), lpid)
	}

	if fds > 1 {
		return 0, fmt.Errorf("multiple file descriptors are not supported, got %d", fds)
	}

	// fd := SD_LISTEN_FDS_START
	// if fd >= fds {
	// 	return 0, fmt.Errorf("no file descriptor available, requested %d, but only %d available", fd, fds)
	// }
	return SdListenFdsStart, nil
}
