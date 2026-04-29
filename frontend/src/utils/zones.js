/**
 * Power zones based on percentage of FTP.
 *
 * Z1 Recovery:    < 55% FTP
 * Z2 Endurance:   55-75% FTP
 * Z3 Tempo:       76-90% FTP
 * Z4 Threshold:   91-105% FTP
 * Z5 VO2max:      106-120% FTP
 * Z6 Anaerobic:   > 120% FTP
 */

export const ZONES = [
  { name: 'Z1', label: 'Recovery', min: 0, max: 0.55, color: '#A0C4E8' },
  { name: 'Z2', label: 'Endurance', min: 0.55, max: 0.75, color: '#3B82F6' },
  { name: 'Z3', label: 'Tempo', min: 0.75, max: 0.90, color: '#22C55E' },
  { name: 'Z4', label: 'Threshold', min: 0.90, max: 1.05, color: '#EAB308' },
  { name: 'Z5', label: 'VO2max', min: 1.05, max: 1.20, color: '#F97316' },
  { name: 'Z6', label: 'Anaerobic', min: 1.20, max: Infinity, color: '#EF4444' },
];

/**
 * Get the zone for a given power as a fraction of FTP.
 * @param {number} powerFraction - Power as fraction of FTP (e.g. 0.75 for 75%)
 * @returns {object} Zone object with name, label, min, max, color
 */
export function getZone(powerFraction) {
  for (const zone of ZONES) {
    if (powerFraction < zone.max) {
      return zone;
    }
  }
  return ZONES[ZONES.length - 1];
}

/**
 * Get the color for a given power as a fraction of FTP.
 * @param {number} powerFraction - Power as fraction of FTP
 * @returns {string} Hex color string
 */
export function getZoneColor(powerFraction) {
  return getZone(powerFraction).color;
}

/**
 * Get zone color for a ramp interval.
 * Uses the average power of start and end for color.
 * @param {number} powerStart - Start power as fraction of FTP
 * @param {number} powerEnd - End power as fraction of FTP
 * @returns {string} Hex color string
 */
export function getRampZoneColor(powerStart, powerEnd) {
  const avg = (powerStart + powerEnd) / 2;
  return getZoneColor(avg);
}

/**
 * Convert a power fraction to absolute watts.
 * @param {number} powerFraction - Power as fraction of FTP
 * @param {number} ftp - User's FTP in watts
 * @returns {number} Power in watts
 */
export function toWatts(powerFraction, ftp) {
  return Math.round(powerFraction * ftp);
}

/**
 * Convert a power fraction to a percentage string.
 * @param {number} powerFraction - Power as fraction of FTP
 * @returns {string} e.g. "75%"
 */
export function toPercent(powerFraction) {
  return `${Math.round(powerFraction * 100)}%`;
}
