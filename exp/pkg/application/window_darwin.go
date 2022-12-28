//go:build darwin

package application

/*
#cgo CFLAGS: -mmacosx-version-min=10.13 -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework WebKit

#include "application.h"
#include "window_delegate.h"
#include <stdlib.h>
#include "Cocoa/Cocoa.h"
#import <WebKit/WebKit.h>
#import <AppKit/AppKit.h>

extern void registerListener(unsigned int event);

// Create a new Window
void* windowNew(unsigned int id, int width, int height) {
	NSWindow* window = [[NSWindow alloc] initWithContentRect:NSMakeRect(0, 0, width-1, height-1)
		styleMask:NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | NSWindowStyleMaskMiniaturizable | NSWindowStyleMaskResizable
		backing:NSBackingStoreBuffered
		defer:NO];

	// Create delegate
	WindowDelegate* delegate = [[WindowDelegate alloc] init];
	// Set delegate
	[window setDelegate:delegate];
	delegate.windowId = id;

	// Add NSView to window
	NSView* view = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, width-1, height-1)];
	[view setAutoresizingMask:NSViewWidthSizable | NSViewHeightSizable];
	[window setContentView:view];

	// Embed wkwebview in window
	NSRect frame = NSMakeRect(0, 0, width, height);
	WKWebViewConfiguration* config = [[WKWebViewConfiguration alloc] init];
	config.suppressesIncrementalRendering = true;
    config.applicationNameForUserAgent = @"wails.io";

	// Setup user content controller
    WKUserContentController* userContentController = [WKUserContentController new];
    [userContentController addScriptMessageHandler:delegate name:@"external"];
    config.userContentController = userContentController;

	WKWebView* webView = [[WKWebView alloc] initWithFrame:frame configuration:config];
	[view addSubview:webView];

    // support webview events
    [webView setNavigationDelegate:delegate];

	// Ensure webview resizes with the window
	[webView setAutoresizingMask:NSViewWidthSizable | NSViewHeightSizable];

	delegate.webView = webView;
	delegate.hideOnClose = false;
	return window;
}

// setInvisibleTitleBarHeight sets the invisible title bar height
void setInvisibleTitleBarHeight(void* window, unsigned int height) {
	NSWindow* nsWindow = (NSWindow*)window;
	// Get delegate
	WindowDelegate* delegate = (WindowDelegate*)[nsWindow delegate];
	// Set height
	delegate.invisibleTitleBarHeight = height;
}

//// Make window toggle frameless
//void windowSetFrameless(void* window, bool frameless) {
//	NSWindow* nsWindow = (NSWindow*)window;
//	if (frameless) {
//		[nsWindow setStyleMask:NSWindowStyleMaskBorderless];
//	} else {
//		[nsWindow setStyleMask:NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | NSWindowStyleMaskMiniaturizable | NSWindowStyleMaskResizable];
//	}
//}

void windowSetFrameless(void* window) {
	NSWindow* nsWindow = (NSWindow*)window;
	[nsWindow setStyleMask:NSWindowStyleMaskBorderless];
}

// Make NSWindow transparent
void windowSetTransparent(void* nsWindow) {
    // On main thread
	dispatch_async(dispatch_get_main_queue(), ^{
	NSWindow* window = (NSWindow*)nsWindow;
	[window setOpaque:NO];
	[window setBackgroundColor:[NSColor clearColor]];
	});
}

void windowSetInvisibleTitleBar(void* nsWindow, unsigned int height) {
	// On main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		NSWindow* window = (NSWindow*)nsWindow;
		// Get delegate
		WindowDelegate* delegate = (WindowDelegate*)[window delegate];
		// Set height
		delegate.invisibleTitleBarHeight = height;
	});
}


// Set the title of the NSWindow
void windowSetTitle(void* nsWindow, char* title) {
	// Set window title on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		NSString* nsTitle = [NSString stringWithUTF8String:title];
		[(NSWindow*)nsWindow setTitle:nsTitle];
		free(title);
	});
}

// Set the size of the NSWindow
void windowSetSize(void* nsWindow, int width, int height) {
	// Set window size on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		NSWindow* window = (NSWindow*)nsWindow;
  		NSSize contentSize = [window contentRectForFrameRect:NSMakeRect(0, 0, width, height)].size;
  		[window setContentSize:contentSize];
  		[window setFrame:NSMakeRect(window.frame.origin.x, window.frame.origin.y, width, height) display:YES animate:YES];
	});
}

// Set NSWindow always on top
void windowSetAlwaysOnTop(void* nsWindow, bool alwaysOnTop) {
	// Set window always on top on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		[(NSWindow*)nsWindow setLevel:alwaysOnTop ? NSStatusWindowLevel : NSNormalWindowLevel];
	});
}

// Load URL in NSWindow
void navigationLoadURL(void* nsWindow, char* url) {
	// Load URL on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		NSURL* nsURL = [NSURL URLWithString:[NSString stringWithUTF8String:url]];
		NSURLRequest* request = [NSURLRequest requestWithURL:nsURL];
		[[(WindowDelegate*)[(NSWindow*)nsWindow delegate] webView] loadRequest:request];
		free(url);
	});
}

// Set NSWindow resizable
void windowSetResizable(void* nsWindow, bool resizable) {
	// Set window resizable on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		NSWindow* window = (NSWindow*)nsWindow;
		if (resizable) {
			[window setStyleMask:[window styleMask] | NSWindowStyleMaskResizable];
		} else {
			[window setStyleMask:[window styleMask] & ~NSWindowStyleMaskResizable];
		}
	});
}

// Set NSWindow min size
void windowSetMinSize(void* nsWindow, int width, int height) {
	// Set window min size on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		NSWindow* window = (NSWindow*)nsWindow;
  		NSSize contentSize = [window contentRectForFrameRect:NSMakeRect(0, 0, width, height)].size;
  		[window setContentMinSize:contentSize];
		NSSize size = { width, height };
  		[window setMinSize:size];
	});
}

// Set NSWindow max size
void windowSetMaxSize(void* nsWindow, int width, int height) {
	// Set window max size on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		NSSize size = { FLT_MAX, FLT_MAX };
    	size.width = width > 0 ? width : FLT_MAX;
    	size.height = height > 0 ? height : FLT_MAX;
		NSWindow* window = (NSWindow*)nsWindow;
  		NSSize contentSize = [window contentRectForFrameRect:NSMakeRect(0, 0, size.width, size.height)].size;
  		[window setContentMaxSize:contentSize];
  		[window setMaxSize:size];
	});
}

// Enable NSWindow devtools
void windowEnableDevTools(void* nsWindow) {
	// Enable devtools on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// Get window delegate
		WindowDelegate* delegate = (WindowDelegate*)[(NSWindow*)nsWindow delegate];
		// Enable devtools in webview
		[delegate.webView.configuration.preferences setValue:@YES forKey:@"developerExtrasEnabled"];
	});
}

// windowResetZoom
void windowResetZoom(void* nsWindow) {
	// Reset zoom on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// Get window delegate
		WindowDelegate* delegate = (WindowDelegate*)[(NSWindow*)nsWindow delegate];
		// Reset zoom
		[delegate.webView setMagnification:1.0];
	});
}

// windowZoomIn
void windowZoomIn(void* nsWindow) {
	// Zoom in on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// Get window delegate
		WindowDelegate* delegate = (WindowDelegate*)[(NSWindow*)nsWindow delegate];
		// Zoom in
		[delegate.webView setMagnification:delegate.webView.magnification + 0.05];
	});
}

// windowZoomOut
void windowZoomOut(void* nsWindow) {
	// Zoom out on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// Get window delegate
		WindowDelegate* delegate = (WindowDelegate*)[(NSWindow*)nsWindow delegate];
		// Zoom out
		if( delegate.webView.magnification > 1.05 ) {
			[delegate.webView setMagnification:delegate.webView.magnification - 0.05];
		} else {
			[delegate.webView setMagnification:1.0];
		}
	});
}

// set the window position
void windowSetPosition(void* nsWindow, int x, int y) {
	// Set window position on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		[(NSWindow*)nsWindow setFrameOrigin:NSMakePoint(x, y)];
	});
}

// Execute JS in NSWindow
void windowExecJS(void* nsWindow, const char* js) {
	// Execute JS on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		WindowDelegate* delegate = (WindowDelegate*)[(NSWindow*)nsWindow delegate];
		[delegate.webView evaluateJavaScript:[NSString stringWithUTF8String:js] completionHandler:nil];
		free((void*)js);
	});
}

// Make NSWindow backdrop translucent
void windowSetTranslucent(void* nsWindow) {
	// Set window transparent on main thread
	dispatch_async(dispatch_get_main_queue(), ^{

		// Get window
		NSWindow* window = (NSWindow*)nsWindow;

		// Get window delegate
		WindowDelegate* delegate = (WindowDelegate*)[(NSWindow*)nsWindow delegate];

		id contentView = [window contentView];
		NSVisualEffectView *effectView = [NSVisualEffectView alloc];
		NSRect bounds = [contentView bounds];
		[effectView initWithFrame:bounds];
		[effectView setAutoresizingMask:NSViewWidthSizable | NSViewHeightSizable];
		[effectView setBlendingMode:NSVisualEffectBlendingModeBehindWindow];
		[effectView setState:NSVisualEffectStateActive];
		[contentView addSubview:effectView positioned:NSWindowBelow relativeTo:nil];
	});
}

// Make webview background transparent
void webviewSetTransparent(void* nsWindow) {
	// Set webview transparent on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// Get window delegate
		WindowDelegate* delegate = (WindowDelegate*)[(NSWindow*)nsWindow delegate];
		// Set webview background transparent
		[delegate.webView setValue:@NO forKey:@"drawsBackground"];
	});
}

// Set webview background colour
void webviewSetBackgroundColour(void* nsWindow, int r, int g, int b, int alpha) {
	// Set webview background color on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// Get window delegate
		WindowDelegate* delegate = (WindowDelegate*)[(NSWindow*)nsWindow delegate];
		// Set webview background color
		[delegate.webView setValue:[NSColor colorWithRed:r/255.0 green:g/255.0 blue:b/255.0 alpha:alpha/255.0] forKey:@"backgroundColor"];
	});
}

// Set the window background colour
void windowSetBackgroundColour(void* nsWindow, int r, int g, int b, int alpha) {
	// Set window background color on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// Get window
		NSWindow* window = (NSWindow*)nsWindow;
		// Set window background color
		[window setBackgroundColor:[NSColor colorWithRed:r/255.0 green:g/255.0 blue:b/255.0 alpha:alpha/255.0]];
	});
}

bool windowIsMaximised(void* nsWindow) {
	return [(NSWindow*)nsWindow isZoomed];
}

bool windowIsFullscreen(void* nsWindow) {
	return [(NSWindow*)nsWindow styleMask] & NSWindowStyleMaskFullScreen;
}

bool windowIsMinimised(void* nsWindow) {
	return [(NSWindow*)nsWindow isMiniaturized];
}

// Set Window fullscreen
void windowFullscreen(void* nsWindow) {
	if( windowIsFullscreen(nsWindow) ) {
		return;
	}
	dispatch_async(dispatch_get_main_queue(), ^{
		[(NSWindow*)nsWindow toggleFullScreen:nil];
	});}

void windowUnFullscreen(void* nsWindow) {
	if( !windowIsFullscreen(nsWindow) ) {
		return;
	}
	dispatch_async(dispatch_get_main_queue(), ^{
		[(NSWindow*)nsWindow toggleFullScreen:nil];
	});
}

// restore window to normal size
void windowRestore(void* nsWindow) {
	// Set window normal on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// If window is fullscreen
		if([(NSWindow*)nsWindow styleMask] & NSWindowStyleMaskFullScreen) {
			[(NSWindow*)nsWindow toggleFullScreen:nil];
		}
		// If window is maximised
		if([(NSWindow*)nsWindow isZoomed]) {
			[(NSWindow*)nsWindow zoom:nil];
		}
		// If window in minimised
		if([(NSWindow*)nsWindow isMiniaturized]) {
			[(NSWindow*)nsWindow deminiaturize:nil];
		}
	});
}

// disable window fullscreen button
void setFullscreenButtonEnabled(void* nsWindow, bool enabled) {
	// Disable fullscreen button on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// Get window
		NSWindow* window = (NSWindow*)nsWindow;
		NSButton *fullscreenButton = [window standardWindowButton:NSWindowZoomButton];
		fullscreenButton.enabled = enabled;
	});
}

// Set the titlebar style
void windowSetTitleBarAppearsTransparent(void* nsWindow, bool transparent) {
	// Set window titlebar style on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		if( transparent ) {
			[(NSWindow*)nsWindow setTitlebarAppearsTransparent:true];
		} else {
			[(NSWindow*)nsWindow setTitlebarAppearsTransparent:false];
		}
	});
}

// Set window fullsize content view
void windowSetFullSizeContent(void* nsWindow, bool fullSize) {
	// Set window fullsize content view on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		if( fullSize ) {
			[(NSWindow*)nsWindow setStyleMask:[(NSWindow*)nsWindow styleMask] | NSWindowStyleMaskFullSizeContentView];
		} else {
			[(NSWindow*)nsWindow setStyleMask:[(NSWindow*)nsWindow styleMask] & ~NSWindowStyleMaskFullSizeContentView];
		}
	});
}

// Set Hide Titlebar
void windowSetHideTitleBar(void* nsWindow, bool hideTitlebar) {
	// Set window titlebar hidden on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		if( hideTitlebar ) {
			[(NSWindow*)nsWindow setStyleMask:[(NSWindow*)nsWindow styleMask] & ~NSWindowStyleMaskTitled];
		} else {
			[(NSWindow*)nsWindow setStyleMask:[(NSWindow*)nsWindow styleMask] | NSWindowStyleMaskTitled];
		}
	});
}

// Set Hide Title in Titlebar
void windowSetHideTitle(void* nsWindow, bool hideTitle) {
	// Set window titlebar hidden on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		if( hideTitle ) {
			[(NSWindow*)nsWindow setTitleVisibility:NSWindowTitleHidden];
		} else {
			[(NSWindow*)nsWindow setTitleVisibility:NSWindowTitleVisible];
		}
	});
}

// Set Window use toolbar
void windowSetUseToolbar(void* nsWindow, bool useToolbar, int toolbarStyle) {
	// Set window use toolbar on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// get main window
		NSWindow* window = (NSWindow*)nsWindow;
		if( useToolbar ) {
			NSToolbar *toolbar = [[NSToolbar alloc] initWithIdentifier:@"wails.toolbar"];
			[toolbar autorelease];
			[window setToolbar:toolbar];

			// If macos 11 or higher, set toolbar style
			if (@available(macOS 11.0, *)) {
				[window setToolbarStyle:toolbarStyle];
			}

		} else {
			[window setToolbar:nil];
		}
	});
}

// Set window toolbar style
void windowSetToolbarStyle(void* nsWindow, int style) {
	// use @available to check if the function is available
	// if not, return
	if (@available(macOS 11.0, *)) {
		// Set window toolbar style on main thread
		dispatch_async(dispatch_get_main_queue(), ^{
			// get main window
			NSWindow* window = (NSWindow*)nsWindow;
			// get toolbar
			NSToolbar* toolbar = [window toolbar];
			// set toolbar style
			[toolbar setShowsBaselineSeparator:style];
		});
	}
}

// Set Hide Toolbar Separator
void windowSetHideToolbarSeparator(void* nsWindow, bool hideSeparator) {
	// Set window hide toolbar separator on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// get main window
		NSWindow* window = (NSWindow*)nsWindow;
		// get toolbar
		NSToolbar* toolbar = [window toolbar];
		// Return if toolbar nil
		if( toolbar == nil ) {
			return;
		}
		if( hideSeparator ) {
			[toolbar setShowsBaselineSeparator:false];
		} else {
			[toolbar setShowsBaselineSeparator:true];
		}
	});
}

// Set Window appearance type
void windowSetAppearanceTypeByName(void* nsWindow, const char *appearanceName) {
	// Set window appearance type on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// get main window
		NSWindow* window = (NSWindow*)nsWindow;
		// set window appearance type by name
		// Convert appearance name to NSString
		NSString* appearanceNameString = [NSString stringWithUTF8String:appearanceName];
		// Set appearance
		[window setAppearance:[NSAppearance appearanceNamed:appearanceNameString]];

		free((void*)appearanceName);
	});
}

// Center window on current monitor
void windowCenter(void* nsWindow) {
	// Center window on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// get main window
		NSWindow* window = (NSWindow*)nsWindow;
		[window center];
	});
}

// Get the current size of the window
void windowGetSize(void* nsWindow, int* width, int* height) {
	// get main window
	NSWindow* window = (NSWindow*)nsWindow;
	// get window frame
	NSRect frame = [window frame];
	// set width and height
	*width = frame.size.width;
	*height = frame.size.height;
}

// Get window width
int windowGetWidth(void* nsWindow) {
	// get main window
	NSWindow* window = (NSWindow*)nsWindow;
	// get window frame
	NSRect frame = [window frame];
	// return width
	return frame.size.width;
}

// Get window height
int windowGetHeight(void* nsWindow) {
	// get main window
	NSWindow* window = (NSWindow*)nsWindow;
	// get window frame
	NSRect frame = [window frame];
	// return height
	return frame.size.height;
}

// Get window position
void windowGetPosition(void* nsWindow, int* x, int* y) {
	// get main window
	NSWindow* window = (NSWindow*)nsWindow;
	// get window frame
	NSRect frame = [window frame];
	// set x and y
	*x = frame.origin.x;
	*y = frame.origin.y;
}

// Destroy window
void windowDestroy(void* nsWindow) {
	// Destroy window on main thread
	dispatch_async(dispatch_get_main_queue(), ^{
		// get main window
		NSWindow* window = (NSWindow*)nsWindow;
		// close window
		[window close];
	});
}


// windowClose closes the current window
static void windowClose(void *window) {
	dispatch_async(dispatch_get_main_queue(), ^{
		// close window
		[(NSWindow*)window close];
	});
}

// windowZoom
static void windowZoom(void *window) {
	dispatch_async(dispatch_get_main_queue(), ^{
		// zoom window
		[(NSWindow*)window zoom:nil];
	});
}

// miniaturize
static void windowMiniaturize(void *window) {
	dispatch_async(dispatch_get_main_queue(), ^{
		// miniaturize window
		[(NSWindow*)window miniaturize:nil];
	});
}

// webviewRenderHTML renders the given HTML
static void windowRenderHTML(void *window, const char *html) {
	dispatch_async(dispatch_get_main_queue(), ^{
		// get main window
		NSWindow* nsWindow = (NSWindow*)window;
		// get window delegate
		WindowDelegate* windowDelegate = (WindowDelegate*)[nsWindow delegate];
		// render html
		[(WKWebView*)windowDelegate.webView loadHTMLString:[NSString stringWithUTF8String:html] baseURL:nil];
	});
}

static void windowInjectCSS(void *window, const char *css) {
	dispatch_async(dispatch_get_main_queue(), ^{
		// get main window
		NSWindow* nsWindow = (NSWindow*)window;
		// get window delegate
		WindowDelegate* windowDelegate = (WindowDelegate*)[nsWindow delegate];
		// inject css
		[(WKWebView*)windowDelegate.webView evaluateJavaScript:[NSString stringWithFormat:@"(function() { var style = document.createElement('style'); style.appendChild(document.createTextNode('%@')); document.head.appendChild(style); })();", [NSString stringWithUTF8String:css]] completionHandler:nil];
        free((void*)css);
	});
}

static void windowMinimise(void *window) {
	dispatch_async(dispatch_get_main_queue(), ^{
		// minimize window
		[(NSWindow*)window miniaturize:nil];
	});
}

// zoom maximizes the window to the screen dimensions
static void windowMaximise(void *window) {
	dispatch_async(dispatch_get_main_queue(), ^{
		// maximize window
		[(NSWindow*)window zoom:nil];
	});
}

static bool isFullScreen(void *window) {
	// get main window
	NSWindow* nsWindow = (NSWindow*)window;
    long mask = [nsWindow styleMask];
    return (mask & NSWindowStyleMaskFullScreen) == NSWindowStyleMaskFullScreen;
}

// windowSetFullScreen
static void windowSetFullScreen(void *window, bool fullscreen) {
	if (isFullScreen(window)) {
		return;
	}
	dispatch_async(dispatch_get_main_queue(), ^{
		NSWindow* nsWindow = (NSWindow*)window;
		windowSetMaxSize(nsWindow, 0, 0);
		windowSetMinSize(nsWindow, 0, 0);
		[nsWindow toggleFullScreen:nil];
	});
}

// windowUnminimise
static void windowUnminimise(void *window) {
	dispatch_async(dispatch_get_main_queue(), ^{
		// unminimize window
		[(NSWindow*)window deminiaturize:nil];
	});
}

// windowUnmaximise
static void windowUnmaximise(void *window) {
	dispatch_async(dispatch_get_main_queue(), ^{
		// unmaximize window
		[(NSWindow*)window zoom:nil];
	});
}

static void windowDisableSizeConstraints(void *window) {
	dispatch_async(dispatch_get_main_queue(), ^{
		// get main window
		NSWindow* nsWindow = (NSWindow*)window;
		// disable size constraints
		[nsWindow setContentMinSize:CGSizeZero];
		[nsWindow setContentMaxSize:CGSizeZero];
	});
}

static void windowShow(void *window) {
	dispatch_async(dispatch_get_main_queue(), ^{
		// show window
		[(NSWindow*)window makeKeyAndOrderFront:nil];
	});
}

static void windowHide(void *window) {
	dispatch_async(dispatch_get_main_queue(), ^{
		[(NSWindow*)window orderOut:nil];
	});
}

*/
import "C"
import (
	"sync"
	"unsafe"

	"github.com/wailsapp/wails/exp/pkg/events"

	"github.com/wailsapp/wails/exp/pkg/options"
)

var showDevTools = func(window unsafe.Pointer) {}

type macosWindow struct {
	nsWindow unsafe.Pointer
	parent   *Window
}

func (w *macosWindow) getScreen() (*Screen, error) {
	return getScreenForWindow(w)
}

func (w *macosWindow) show() {
	C.windowShow(w.nsWindow)
}

func (w *macosWindow) hide() {
	C.windowHide(w.nsWindow)
}

func (w *macosWindow) setFullscreenButtonEnabled(enabled bool) {
	C.setFullscreenButtonEnabled(w.nsWindow, C.bool(enabled))
}

func (w *macosWindow) disableSizeConstraints() {
	C.windowDisableSizeConstraints(w.nsWindow)
}

func (w *macosWindow) unfullscreen() {
	C.windowUnFullscreen(w.nsWindow)
}

func (w *macosWindow) fullscreen() {
	C.windowFullscreen(w.nsWindow)
}

func (w *macosWindow) unminimise() {
	C.windowUnminimise(w.nsWindow)
}

func (w *macosWindow) unmaximise() {
	C.windowUnmaximise(w.nsWindow)
}

func (w *macosWindow) maximise() {
	C.windowMaximise(w.nsWindow)
}

func (w *macosWindow) minimise() {
	C.windowMinimise(w.nsWindow)
}

func (w *macosWindow) on(eventID uint) {
	C.registerListener(C.uint(eventID))
}

func (w *macosWindow) zoom() {
	C.windowZoom(w.nsWindow)
}

func (w *macosWindow) minimize() {
	C.windowMiniaturize(w.nsWindow)
}

func (w *macosWindow) windowZoom() {
	C.windowZoom(w.nsWindow)
}

func (w *macosWindow) close() {
	C.windowClose(w.nsWindow)
}

func (w *macosWindow) zoomIn() {
	C.windowZoomIn(w.nsWindow)
}

func (w *macosWindow) zoomOut() {
	C.windowZoomOut(w.nsWindow)
}

func (w *macosWindow) resetZoom() {
	C.windowResetZoom(w.nsWindow)
}

func (w *macosWindow) toggleDevTools() {
	showDevTools(w.nsWindow)
}

func (w *macosWindow) reload() {
	//TODO: Implement
	println("reload called on Window", w.parent.id)
}

func (w *macosWindow) forceReload() {
	//TODO: Implement
	println("forceReload called on Window", w.parent.id)
}

func (w *macosWindow) center() {
	C.windowCenter(w.nsWindow)
}

func (w *macosWindow) isMinimised() bool {
	return w.syncMainThreadReturningBool(func() bool {
		return bool(C.windowIsMinimised(w.nsWindow))
	})
}

func (w *macosWindow) isMaximised() bool {
	return w.syncMainThreadReturningBool(func() bool {
		return bool(C.windowIsMaximised(w.nsWindow))
	})
}

func (w *macosWindow) isFullscreen() bool {
	return w.syncMainThreadReturningBool(func() bool {
		return bool(C.windowIsFullscreen(w.nsWindow))
	})
}

func (w *macosWindow) syncMainThreadReturningBool(fn func() bool) bool {
	var wg sync.WaitGroup
	wg.Add(1)
	var result bool
	globalApplication.dispatchOnMainThread(func() {
		result = fn()
		wg.Done()
	})
	wg.Wait()
	return result
}

func (w *macosWindow) restore() {
	// restore window to normal size
	C.windowRestore(w.nsWindow)
}

func (w *macosWindow) restoreWindow() {
	C.windowRestore(w.nsWindow)
}

func (w *macosWindow) execJS(js string) {
	println("execJS called on Window", w.parent.id)
	C.windowExecJS(w.nsWindow, C.CString(js))
}

func (w *macosWindow) setURL(url string) {
	C.navigationLoadURL(w.nsWindow, C.CString(url))
}

func (w *macosWindow) setAlwaysOnTop(alwaysOnTop bool) {
	C.windowSetAlwaysOnTop(w.nsWindow, C.bool(alwaysOnTop))
}

func newWindowImpl(parent *Window) *macosWindow {
	result := &macosWindow{
		parent: parent,
	}
	return result
}

func (w *macosWindow) setTitle(title string) {
	cTitle := C.CString(title)
	C.windowSetTitle(w.nsWindow, cTitle)
}

func (w *macosWindow) setSize(width, height int) {
	C.windowSetSize(w.nsWindow, C.int(width), C.int(height))
}

func (w *macosWindow) setMinSize(width, height int) {
	C.windowSetMinSize(w.nsWindow, C.int(width), C.int(height))
}
func (w *macosWindow) setMaxSize(width, height int) {
	C.windowSetMaxSize(w.nsWindow, C.int(width), C.int(height))
}

func (w *macosWindow) setResizable(resizable bool) {
	C.windowSetResizable(w.nsWindow, C.bool(resizable))
}
func (w *macosWindow) enableDevTools() {
	C.windowEnableDevTools(w.nsWindow)
}

func (w *macosWindow) size() (int, int) {
	var width, height C.int
	var wg sync.WaitGroup
	wg.Add(1)
	globalApplication.dispatchOnMainThread(func() {
		C.windowGetSize(w.nsWindow, &width, &height)
		wg.Done()
	})
	wg.Wait()
	return int(width), int(height)
}

func (w *macosWindow) setPosition(x, y int) {
	C.windowSetPosition(w.nsWindow, C.int(x), C.int(y))
}

func (w *macosWindow) width() int {
	var width C.int
	var wg sync.WaitGroup
	wg.Add(1)
	globalApplication.dispatchOnMainThread(func() {
		width = C.windowGetWidth(w.nsWindow)
		wg.Done()
	})
	wg.Wait()
	return int(width)
}
func (w *macosWindow) height() int {
	var height C.int
	var wg sync.WaitGroup
	wg.Add(1)
	globalApplication.dispatchOnMainThread(func() {
		height = C.windowGetHeight(w.nsWindow)
		wg.Done()
	})
	wg.Wait()
	return int(height)
}

func (w *macosWindow) run() {
	for eventId := range w.parent.eventListeners {
		w.on(eventId)
	}
	globalApplication.dispatchOnMainThread(func() {
		w.nsWindow = C.windowNew(C.uint(w.parent.id), C.int(w.parent.options.Width), C.int(w.parent.options.Height))
		w.setTitle(w.parent.options.Title)
		w.setAlwaysOnTop(w.parent.options.AlwaysOnTop)
		w.setResizable(!w.parent.options.DisableResize)
		if w.parent.options.MinWidth != 0 || w.parent.options.MinHeight != 0 {
			w.setMinSize(w.parent.options.MinWidth, w.parent.options.MinHeight)
		}
		if w.parent.options.MaxWidth != 0 || w.parent.options.MaxHeight != 0 {
			w.setMaxSize(w.parent.options.MaxWidth, w.parent.options.MaxHeight)
		}
		w.enableDevTools()
		w.setBackgroundColour(w.parent.options.BackgroundColour)
		if w.parent.options.Mac != nil {
			macOptions := w.parent.options.Mac
			switch macOptions.Backdrop {
			case options.MacBackdropTransparent:
				C.windowSetTransparent(w.nsWindow)
				C.webviewSetTransparent(w.nsWindow)
			case options.MacBackdropTranslucent:
				C.windowSetTranslucent(w.nsWindow)
				C.webviewSetTransparent(w.nsWindow)
			}

			if macOptions.TitleBar != nil {
				titleBarOptions := macOptions.TitleBar
				C.windowSetTitleBarAppearsTransparent(w.nsWindow, C.bool(titleBarOptions.AppearsTransparent))
				C.windowSetHideTitleBar(w.nsWindow, C.bool(titleBarOptions.Hide))
				C.windowSetHideTitle(w.nsWindow, C.bool(titleBarOptions.HideTitle))
				C.windowSetFullSizeContent(w.nsWindow, C.bool(titleBarOptions.FullSizeContent))
				if titleBarOptions.UseToolbar {
					C.windowSetUseToolbar(w.nsWindow, C.bool(titleBarOptions.UseToolbar), C.int(titleBarOptions.ToolbarStyle))
				}
				C.windowSetHideToolbarSeparator(w.nsWindow, C.bool(titleBarOptions.HideToolbarSeparator))
			}

			if macOptions.Appearance != "" {
				C.windowSetAppearanceTypeByName(w.nsWindow, C.CString(string(macOptions.Appearance)))
			}

			if macOptions.InvisibleTitleBarHeight != 0 {
				C.windowSetInvisibleTitleBar(w.nsWindow, C.uint(macOptions.InvisibleTitleBarHeight))
			}
		}

		switch w.parent.options.StartState {
		case options.WindowStateMaximised:
			w.maximise()
		case options.WindowStateMinimised:
			w.minimise()
		case options.WindowStateFullscreen:
			w.fullscreen()

		}
		C.windowCenter(w.nsWindow)

		if w.parent.options.URL != "" {
			w.setURL(w.parent.options.URL)
		}
		// We need to wait for the HTML to load before we can execute the javascript
		w.parent.On(events.Mac.WebViewDidFinishNavigation, func() {
			if w.parent.options.JS != "" {
				w.execJS(w.parent.options.JS)
			}
			if w.parent.options.CSS != "" {
				C.windowInjectCSS(w.nsWindow, C.CString(w.parent.options.CSS))
			}
		})
		if w.parent.options.HTML != "" {
			w.setHTML(w.parent.options.HTML)
		}
		if w.parent.options.Hidden == false {
			C.windowShow(w.nsWindow)
		}
	})
}

func (w *macosWindow) setBackgroundColour(colour *options.RGBA) {
	if colour == nil {
		return
	}
	C.windowSetBackgroundColour(w.nsWindow, C.int(colour.Red), C.int(colour.Green), C.int(colour.Blue), C.int(colour.Alpha))
}

func (w *macosWindow) position() (int, int) {
	var x, y C.int
	var wg sync.WaitGroup
	wg.Add(1)
	go globalApplication.dispatchOnMainThread(func() {
		C.windowGetPosition(w.nsWindow, &x, &y)
		wg.Done()
	})
	wg.Wait()
	return int(x), int(y)
}

func (w *macosWindow) destroy() {
	C.windowDestroy(w.nsWindow)
}

func (w *macosWindow) setHTML(html string) {
	// Convert HTML to C string
	cHTML := C.CString(html)
	// Render HTML
	C.windowRenderHTML(w.nsWindow, cHTML)
}