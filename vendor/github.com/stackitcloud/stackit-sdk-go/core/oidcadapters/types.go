package oidcadapters

import "context"

type OIDCTokenFunc func(context.Context) (string, error)
