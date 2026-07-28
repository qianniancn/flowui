//go:build darwin

package darwin

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#import <Foundation/Foundation.h>

const char* getPreferredLanguage() {
	NSArray *languages = [NSLocale preferredLanguages];
	if (languages.count == 0) {
		return "";
	}
	NSString *language = languages[0];
	return [language UTF8String];
}
*/
import "C"

// SystemLocaleName returns the system locale name on macOS.
// It uses NSLocale.preferredLanguages to get the user's preferred language.
func SystemLocaleName() string {
	cStr := C.getPreferredLanguage()
	return C.GoString(cStr)
}
