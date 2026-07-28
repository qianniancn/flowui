//go:build darwin && !ios

#ifndef SYSTRAY_H
#define SYSTRAY_H

#include <Cocoa/Cocoa.h>

@interface StatusItemController : NSObject <NSMenuDelegate>
@property long id;
@property (assign) NSStatusItem *statusItem;
@property (assign) NSMenu *cachedMenu;
@property (strong) id eventMonitor;
- (void)statusItemClicked:(id)sender;
@end

// System tray functions
void* systemTrayNew(long id);
void systemTraySetLabel(void* nsStatusItem, char *label);
NSImage* imageFromBytes(const unsigned char *bytes, int length);
void systemTraySetIcon(void* nsStatusItem, void* nsImage, int position, bool isTemplate);
void systemTrayDestroy(void* nsStatusItem);
void systemTrayShow(void* nsStatusItem);
void systemTrayHide(void* nsStatusItem);
void showMenu(void* nsStatusItem, void *nsMenu);
void systemTraySetCachedMenu(void* nsStatusItem, void *nsMenu);

// Menu functions
void* createNSMenu(char* label);
void releaseNSMenu(void* nsMenu);
void addMenuItem(void* nsMenu, void* nsMenuItem);
void addMenuSeparator(void* nsMenu);
void setMenuItemSubmenu(void* nsMenuItem, void* nsMenu);
void clearMenu(void* nsMenu);

// Menu item functions
void* createNSMenuItem(char* label, long id);
void setMenuItemChecked(void* nsMenuItem, bool checked);
void setMenuItemEnabled(void* nsMenuItem, bool enabled);
void setMenuItemHidden(void* nsMenuItem, bool hidden);
void releaseNSMenuItem(void* nsMenuItem);

// Callbacks - implemented in Go
extern void systrayClickCallback(long id, int buttonID);
extern int systrayPreClickCallback(long id, int buttonID);
extern void menuItemCallback(long id);

#endif /* SYSTRAY_H */
