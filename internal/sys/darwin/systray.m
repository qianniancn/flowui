//go:build darwin && !ios

#include "Cocoa/Cocoa.h"
#include "systray.h"

extern void systrayClickCallback(long, int);
extern int systrayPreClickCallback(long, int);
extern void menuItemCallback(long);

static void runOnMainSync(void (^block)(void)) {
    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_sync(dispatch_get_main_queue(), block);
    }
}

// StatusItemController implementation
@implementation StatusItemController

- (void)statusItemClicked:(id)sender {
    NSEvent *event = [NSApp currentEvent];
    systrayClickCallback(self.id, event.type);
}

- (void)menuDidClose:(NSMenu *)menu {
    // Remove the menu from the status item so future clicks invoke the
    // action handler instead of re-showing the menu.
    self.statusItem.menu = nil;
    menu.delegate = nil;
}

@end

// Create a new system tray
void* systemTrayNew(long id) {
    __block NSStatusItem *result = nil;
    runOnMainSync(^{
        StatusItemController *controller = [[StatusItemController alloc] init];
        controller.id = id;
        NSStatusItem *statusItem = [[[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength] retain];
        controller.statusItem = statusItem;
        [statusItem setTarget:controller];
        [statusItem setAction:@selector(statusItemClicked:)];
        NSButton *button = statusItem.button;
        [button sendActionOn:(NSEventMaskLeftMouseDown|NSEventMaskRightMouseDown)];

        controller.eventMonitor = [NSEvent addLocalMonitorForEventsMatchingMask:
            (NSEventMaskLeftMouseDown|NSEventMaskRightMouseDown)
            handler:^NSEvent *(NSEvent *event) {
                if (event.window != button.window) return event;

                int action = systrayPreClickCallback((long)controller.id, (int)event.type);
                if (action == 1 && controller.cachedMenu != nil) {
                    controller.cachedMenu.delegate = controller;
                    statusItem.menu = controller.cachedMenu;
                }
                return event;
            }];
        result = statusItem;
    });
    return (void*)result;
}

void systemTraySetLabel(void* nsStatusItem, char *label) {
    if(label == NULL) {
        return;
    }
    runOnMainSync(^{
        NSStatusItem *statusItem = (NSStatusItem *)nsStatusItem;
        statusItem.button.title = [NSString stringWithUTF8String:label];
        free(label);
    });
}

// Create an NSImage from a byte array
NSImage* imageFromBytes(const unsigned char *bytes, int length) {
    __block NSImage *image = nil;
    runOnMainSync(^{
        NSData *data = [NSData dataWithBytes:bytes length:length];
        image = [[NSImage alloc] initWithData:data];
    });
    return image;
}

// Set the icon on the system tray
void systemTraySetIcon(void* nsStatusItem, void* nsImage, int position, bool isTemplate) {
    runOnMainSync(^{
        NSStatusItem *statusItem = (NSStatusItem *)nsStatusItem;
        NSImage *image = (NSImage *)nsImage;

        NSStatusBar *statusBar = [NSStatusBar systemStatusBar];
        CGFloat thickness = [statusBar thickness];
        [image setSize:NSMakeSize(thickness, thickness)];
        if(isTemplate) {
            [image setTemplate:YES];
        }
        statusItem.button.image = [image autorelease];
        statusItem.button.imagePosition = position;
    });
}

// Destroy system tray
void systemTrayDestroy(void* nsStatusItem) {
    // Remove the status item from the status bar and its associated menu
    runOnMainSync(^{
        NSStatusItem *statusItem = (NSStatusItem *)nsStatusItem;
        StatusItemController *controller = (StatusItemController *)[statusItem target];
        if (controller.eventMonitor) {
            [NSEvent removeMonitor:controller.eventMonitor];
            controller.eventMonitor = nil;
        }
        [[NSStatusBar systemStatusBar] removeStatusItem:statusItem];
        [controller release];
        [statusItem release];
    });
}

// Show the system tray icon
void systemTrayShow(void* nsStatusItem) {
    runOnMainSync(^{
        NSStatusItem *statusItem = (NSStatusItem *)nsStatusItem;
        [statusItem setVisible:YES];
    });
}

// Hide the system tray icon
void systemTrayHide(void* nsStatusItem) {
    runOnMainSync(^{
        NSStatusItem *statusItem = (NSStatusItem *)nsStatusItem;
        [statusItem setVisible:NO];
    });
}

// showMenu is used for programmatic OpenMenu() calls. Click-triggered
// menus are handled by the event monitor installed in systemTrayNew.
void showMenu(void* nsStatusItem, void *nsMenu) {
    runOnMainSync(^{
        NSStatusItem *statusItem = (NSStatusItem *)nsStatusItem;
        NSMenu *menu = (NSMenu *)nsMenu;
        StatusItemController *controller = (StatusItemController *)[statusItem target];

        // Temporarily assign the menu for native tracking.
        menu.delegate = controller;
        statusItem.menu = menu;

        // Synthesize a mouse-down at the button centre to trigger native
        // menu tracking (highlights the button, blocks until dismissed).
        NSRect frame = [statusItem.button convertRect:statusItem.button.bounds toView:nil];
        NSPoint loc = NSMakePoint(NSMidX(frame), NSMidY(frame));
        NSEvent *event = [NSEvent mouseEventWithType:NSEventTypeLeftMouseDown
                                            location:loc
                                       modifierFlags:0
                                           timestamp:[[NSProcessInfo processInfo] systemUptime]
                                        windowNumber:statusItem.button.window.windowNumber
                                             context:nil
                                         eventNumber:0
                                          clickCount:1
                                            pressure:1.0];
        [statusItem.button mouseDown:event];

        // Menu dismissed — restore custom click handling.
        statusItem.menu = nil;
        menu.delegate = nil;
    });
}

void systemTraySetCachedMenu(void* nsStatusItem, void *nsMenu) {
    runOnMainSync(^{
        NSStatusItem *statusItem = (NSStatusItem *)nsStatusItem;
        StatusItemController *controller = (StatusItemController *)[statusItem target];
        controller.cachedMenu = (NSMenu *)nsMenu;
    });
}

// Create a new NSMenu
void* createNSMenu(char* label) {
    __block NSMenu *menu = nil;
    runOnMainSync(^{
        menu = [[NSMenu alloc] init];
        if(label != NULL && strlen(label) > 0) {
            menu.title = [NSString stringWithUTF8String:label];
        }
        if(label != NULL) {
            free(label);
        }
        [menu setAutoenablesItems:NO];
    });
    return (void*)menu;
}

// Release a menu created by createNSMenu
void releaseNSMenu(void* nsMenu) {
    runOnMainSync(^{
        NSMenu *menu = (NSMenu *)nsMenu;
        [menu release];
    });
}

void addMenuItem(void* nsMenu, void* nsMenuItem) {
    runOnMainSync(^{
        NSMenu *menu = (NSMenu *)nsMenu;
        [menu addItem:nsMenuItem];
    });
}

// Add separator to menu
void addMenuSeparator(void* nsMenu) {
    runOnMainSync(^{
        NSMenu *menu = (NSMenu *)nsMenu;
        [menu addItem:[NSMenuItem separatorItem]];
    });
}

// Set the submenu of a menu item
void setMenuItemSubmenu(void* nsMenuItem, void* nsMenu) {
    runOnMainSync(^{
        NSMenuItem *menuItem = (NSMenuItem *)nsMenuItem;
        NSMenu *menu = (NSMenu *)nsMenu;
        [menuItem setSubmenu:menu];
    });
}

// Clear and release all menu items in the menu
void clearMenu(void* nsMenu) {
    runOnMainSync(^{
        NSMenu *menu = (NSMenu *)nsMenu;
        [menu removeAllItems];
    });
}

// Create a new NSMenuItem
void* createNSMenuItem(char* label, long id) {
    __block NSMenuItem *menuItem = nil;
    runOnMainSync(^{
        menuItem = [[NSMenuItem alloc] init];
        if(label != NULL) {
            menuItem.title = [NSString stringWithUTF8String:label];
            free(label);
        }
        menuItem.target = nil;
        menuItem.action = @selector(menuItemClicked:);
        menuItem.tag = id;
    });
    return (void*)menuItem;
}

void setMenuItemChecked(void* nsMenuItem, bool checked) {
    runOnMainSync(^{
        NSMenuItem *menuItem = (NSMenuItem *)nsMenuItem;
        if(checked) {
            [menuItem setState:NSControlStateValueOn];
        } else {
            [menuItem setState:NSControlStateValueOff];
        }
    });
}

void setMenuItemEnabled(void* nsMenuItem, bool enabled) {
    runOnMainSync(^{
        NSMenuItem *menuItem = (NSMenuItem *)nsMenuItem;
        [menuItem setEnabled:enabled];
    });
}

void setMenuItemHidden(void* nsMenuItem, bool hidden) {
    runOnMainSync(^{
        NSMenuItem *menuItem = (NSMenuItem *)nsMenuItem;
        [menuItem setHidden:hidden];
    });
}

void releaseNSMenuItem(void* nsMenuItem) {
    runOnMainSync(^{
        NSMenuItem *menuItem = (NSMenuItem *)nsMenuItem;
        [menuItem release];
    });
}
