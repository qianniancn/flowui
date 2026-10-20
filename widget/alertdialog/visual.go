package alertdialog

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type alertDialogVisualKey struct {
	status      Status
	size        Size
	backdrop    BackdropVariant
	override    uint64
	overrideSet bool
}

type alertDialogVisualCache struct {
	cache computed.ProviderCache[alertDialogVisualKey, theme.AlertDialogVisual]
}

func (a Widget) visualWithCache(ctx *core.Context, size Size, backdrop BackdropVariant, status Status, override *theme.ModalOverride) theme.AlertDialogVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	hash, set := alertDialogOverrideKey(override)
	key := alertDialogVisualKey{
		status: status, size: size, backdrop: backdrop,
		override: hash, overrideSet: set,
	}
	load := func() theme.AlertDialogVisual {
		return theme.ResolveVisual(active, theme.AlertDialogRequest{
			Status: theme.AlertStatus(status), Size: theme.ModalSize(a.modalSize()),
			Backdrop: theme.ModalBackdrop(a.modalBackdrop()), Override: override,
		}, theme.ComponentVisualProvider.AlertDialog, theme.ValidateAlertDialogVisual)
	}
	if a.visualCache == nil {
		return load()
	}
	return a.visualCache.cache.Resolve(active, key, load)
}

func alertDialogOverrideKey(override *theme.ModalOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	return override.Declaration.Hash64(), true
}

func (h dialogHeader) visual(ctx *core.Context) theme.AlertDialogVisual {
	return Widget{status: h.status, size: h.size, backdrop: h.backdrop, visualCache: h.cache}.
		visualWithCache(ctx, h.size, h.backdrop, h.status, h.override)
}
