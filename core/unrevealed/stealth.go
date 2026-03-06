// stealth.go
package unrevealed

// StealthScripts returns JavaScript snippets to inject via CDP
// Page.addScriptToEvaluateOnNewDocument for bot detection evasion.
func StealthScripts() []string {
	return []string{
		scriptWebdriver,
		scriptChrome,
		scriptPermissions,
		scriptFnToString,
		scriptWebGL,
		scriptHardware,
		scriptCanvas,
	}
}

// StealthFlags returns Chrome command-line flags for stealth operation.
// Keys are flag names, values are flag values (empty string for boolean flags).
func StealthFlags() map[string]string {
	return map[string]string{
		"disable-blink-features":   "AutomationControlled",
		"no-first-run":             "",
		"no-default-browser-check": "",
		"start-maximized":          "",
		"disable-infobars":         "",
	}
}

// DeleteFlags returns Chrome flags that should be removed from launch
// arguments to avoid detection (typically added by automation libraries).
func DeleteFlags() []string {
	return []string{
		"enable-automation",
	}
}

// Override navigator.webdriver on the prototype.
// The --disable-blink-features=AutomationControlled flag handles this at the
// Chrome level, but this JS override acts as a fallback for edge cases.
var scriptWebdriver = `
Object.defineProperty(Navigator.prototype, 'webdriver', {
  get: () => undefined,
  configurable: true,
});
`

// Mock window.chrome.runtime which is absent in automated browsers
// but present in regular Chrome sessions.
var scriptChrome = `
if (!window.chrome) {
  window.chrome = {};
}
if (!window.chrome.runtime) {
  window.chrome.runtime = {
    OnInstalledReason: {
      CHROME_UPDATE: 'chrome_update',
      INSTALL: 'install',
      SHARED_MODULE_UPDATE: 'shared_module_update',
      UPDATE: 'update',
    },
    OnRestartRequiredReason: {
      APP_UPDATE: 'app_update',
      OS_UPDATE: 'os_update',
      PERIODIC: 'periodic',
    },
    PlatformArch: {
      ARM: 'arm',
      ARM64: 'arm64',
      MIPS: 'mips',
      MIPS64: 'mips64',
      X86_32: 'x86-32',
      X86_64: 'x86-64',
    },
    PlatformNaclArch: {
      ARM: 'arm',
      MIPS: 'mips',
      MIPS64: 'mips64',
      X86_32: 'x86-32',
      X86_64: 'x86-64',
    },
    PlatformOs: {
      ANDROID: 'android',
      CROS: 'cros',
      LINUX: 'linux',
      MAC: 'mac',
      OPENBSD: 'openbsd',
      WIN: 'win',
    },
    RequestUpdateCheckStatus: {
      NO_UPDATE: 'no_update',
      THROTTLED: 'throttled',
      UPDATE_AVAILABLE: 'update_available',
    },
  };
}
`

// Fix navigator.permissions.query behavior and Notification API
// which differ in automated browsers.
var scriptPermissions = `
if (!window.Notification) {
  window.Notification = { permission: 'denied' };
}
const originalQuery = window.navigator.permissions.query;
window.navigator.permissions.__proto__.query = (parameters) =>
  parameters.name === 'notifications'
    ? Promise.resolve({ state: Notification.permission })
    : originalQuery(parameters);
`

// Spoof Function.prototype.toString so overridden native functions
// still return "[native code]" when inspected by detection scripts.
var scriptFnToString = `
(() => {
  const nativeStr = Error.toString().replace(/Error/g, 'toString');
  const oldToString = Function.prototype.toString;
  const oldCall = Function.prototype.call;

  function call() {
    return oldCall.apply(this, arguments);
  }
  Function.prototype.call = call;

  function functionToString() {
    if (this === window.navigator.permissions.query) {
      return 'function query() { [native code] }';
    }
    if (this === functionToString) {
      return nativeStr;
    }
    return oldCall.call(oldToString, this);
  }
  Function.prototype.toString = functionToString;
})();
`

// Spoof WebGL renderer and vendor strings so detection scripts cannot
// identify the GPU as a virtual or headless device.
var scriptWebGL = `
(() => {
  const getParam = WebGLRenderingContext.prototype.getParameter;
  WebGLRenderingContext.prototype.getParameter = function(p) {
    if (p === 37445) return 'Google Inc. (Intel)';
    if (p === 37446) return 'ANGLE (Intel, Intel(R) UHD Graphics 630, OpenGL 4.5)';
    return getParam.call(this, p);
  };
  if (typeof WebGL2RenderingContext !== 'undefined') {
    const getParam2 = WebGL2RenderingContext.prototype.getParameter;
    WebGL2RenderingContext.prototype.getParameter = function(p) {
      if (p === 37445) return 'Google Inc. (Intel)';
      if (p === 37446) return 'ANGLE (Intel, Intel(R) UHD Graphics 630, OpenGL 4.5)';
      return getParam2.call(this, p);
    };
  }
})();
`

// Set realistic values for hardware-related navigator properties
// that differ in headless or automated environments.
var scriptHardware = `
Object.defineProperty(navigator, 'hardwareConcurrency', { get: () => 8 });
Object.defineProperty(navigator, 'deviceMemory', { get: () => 8 });
Object.defineProperty(navigator, 'maxTouchPoints', { get: () => 0 });
Object.defineProperty(navigator, 'languages', { get: () => ['en-US', 'en'] });
if (navigator.connection) {
  Object.defineProperty(navigator.connection, 'rtt', { get: () => 50 });
}
`

// Add subtle noise to canvas toDataURL and toBlob to defeat
// canvas fingerprinting without visually breaking pages.
var scriptCanvas = `
(() => {
  const origToDataURL = HTMLCanvasElement.prototype.toDataURL;
  HTMLCanvasElement.prototype.toDataURL = function(type) {
    const ctx = this.getContext('2d');
    if (ctx) {
      const s = ctx.fillStyle;
      ctx.fillStyle = 'rgba(0,0,1,0.003)';
      ctx.fillRect(0, 0, 1, 1);
      ctx.fillStyle = s;
    }
    return origToDataURL.apply(this, arguments);
  };
  const origToBlob = HTMLCanvasElement.prototype.toBlob;
  HTMLCanvasElement.prototype.toBlob = function(cb, type, quality) {
    const ctx = this.getContext('2d');
    if (ctx) {
      const s = ctx.fillStyle;
      ctx.fillStyle = 'rgba(0,0,1,0.003)';
      ctx.fillRect(0, 0, 1, 1);
      ctx.fillStyle = s;
    }
    return origToBlob.apply(this, arguments);
  };
})();
`
