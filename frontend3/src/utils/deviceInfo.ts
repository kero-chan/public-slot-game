/**
 * Collects device and browser information for session tracking
 */

interface DeviceInfoData {
  browser: string
  browserVersion: string
  os: string
  osVersion: string
  device: string
  screenResolution: string
  language: string
  timezone: string
}

/**
 * Parse browser and version from user agent
 */
function getBrowserInfo(): { browser: string; version: string } {
  const ua = navigator.userAgent
  let browser = 'Unknown'
  let version = ''

  if (ua.includes('Firefox/')) {
    browser = 'Firefox'
    version = ua.match(/Firefox\/(\d+\.?\d*)/)?.[1] || ''
  } else if (ua.includes('Edg/')) {
    browser = 'Edge'
    version = ua.match(/Edg\/(\d+\.?\d*)/)?.[1] || ''
  } else if (ua.includes('Chrome/')) {
    browser = 'Chrome'
    version = ua.match(/Chrome\/(\d+\.?\d*)/)?.[1] || ''
  } else if (ua.includes('Safari/') && !ua.includes('Chrome')) {
    browser = 'Safari'
    version = ua.match(/Version\/(\d+\.?\d*)/)?.[1] || ''
  } else if (ua.includes('Opera') || ua.includes('OPR/')) {
    browser = 'Opera'
    version = ua.match(/(?:Opera|OPR)\/(\d+\.?\d*)/)?.[1] || ''
  }

  return { browser, version }
}

/**
 * Parse OS and version from user agent
 */
function getOSInfo(): { os: string; version: string } {
  const ua = navigator.userAgent
  let os = 'Unknown'
  let version = ''

  if (ua.includes('Windows NT')) {
    os = 'Windows'
    const ntVersion = ua.match(/Windows NT (\d+\.?\d*)/)?.[1]
    if (ntVersion === '10.0') version = '10/11'
    else if (ntVersion === '6.3') version = '8.1'
    else if (ntVersion === '6.2') version = '8'
    else if (ntVersion === '6.1') version = '7'
    else version = ntVersion || ''
  } else if (ua.includes('Mac OS X')) {
    os = 'macOS'
    version = ua.match(/Mac OS X (\d+[._]\d+)/)?.[1]?.replace('_', '.') || ''
  } else if (ua.includes('Android')) {
    os = 'Android'
    version = ua.match(/Android (\d+\.?\d*)/)?.[1] || ''
  } else if (ua.includes('iPhone') || ua.includes('iPad')) {
    os = 'iOS'
    version = ua.match(/OS (\d+[._]\d+)/)?.[1]?.replace('_', '.') || ''
  } else if (ua.includes('Linux')) {
    os = 'Linux'
  }

  return { os, version }
}

/**
 * Detect device type
 */
function getDeviceType(): string {
  const ua = navigator.userAgent

  if (ua.includes('iPad')) return 'Tablet'
  if (ua.includes('iPhone')) return 'Mobile'
  if (ua.includes('Android')) {
    if (ua.includes('Mobile')) return 'Mobile'
    return 'Tablet'
  }
  if (ua.includes('Mobile')) return 'Mobile'

  return 'Desktop'
}

/**
 * Collect all device information
 */
export function collectDeviceInfo(): DeviceInfoData {
  const { browser, version: browserVersion } = getBrowserInfo()
  const { os, version: osVersion } = getOSInfo()

  return {
    browser,
    browserVersion,
    os,
    osVersion,
    device: getDeviceType(),
    screenResolution: `${window.screen.width}x${window.screen.height}`,
    language: navigator.language,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  }
}

/**
 * Get device info as a formatted string for API
 * Format: "Browser/Version (OS Version; Device; Resolution)"
 */
export function getDeviceInfoString(): string {
  const info = collectDeviceInfo()

  const parts = [
    `${info.browser}/${info.browserVersion}`,
    `(${info.os}${info.osVersion ? ' ' + info.osVersion : ''}`,
    `${info.device}`,
    `${info.screenResolution})`,
  ]

  return parts.join('; ')
}
