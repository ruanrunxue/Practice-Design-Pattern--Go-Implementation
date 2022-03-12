package plugin

import "errors"

var (
	ErrPluginNotInstalled = errors.New("plugin not installed yet")
	ErrPluginUninstalled  = errors.New("plugin uninstalled")
	ErrUnknownPlugin      = errors.New("unknown plugin")
)
