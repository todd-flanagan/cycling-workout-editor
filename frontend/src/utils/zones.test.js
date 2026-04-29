import { describe, it, expect } from 'vitest';
import { ZONES, getZone, getZoneColor, getRampZoneColor, toWatts, toPercent } from './zones';

describe('zones', () => {
  describe('ZONES constant', () => {
    it('has 6 zones', () => {
      expect(ZONES).toHaveLength(6);
    });

    it('zones are ordered by min value', () => {
      for (let i = 1; i < ZONES.length; i++) {
        expect(ZONES[i].min).toBeGreaterThanOrEqual(ZONES[i - 1].min);
      }
    });

    it('each zone has required fields', () => {
      ZONES.forEach((zone) => {
        expect(zone).toHaveProperty('name');
        expect(zone).toHaveProperty('label');
        expect(zone).toHaveProperty('min');
        expect(zone).toHaveProperty('max');
        expect(zone).toHaveProperty('color');
        expect(zone.color).toMatch(/^#[0-9A-Fa-f]{6}$/);
      });
    });
  });

  describe('getZone', () => {
    it('returns Z1 for power < 55%', () => {
      expect(getZone(0.40).name).toBe('Z1');
      expect(getZone(0.54).name).toBe('Z1');
      expect(getZone(0.0).name).toBe('Z1');
    });

    it('returns Z2 for power 55-75%', () => {
      expect(getZone(0.55).name).toBe('Z2');
      expect(getZone(0.65).name).toBe('Z2');
      expect(getZone(0.74).name).toBe('Z2');
    });

    it('returns Z3 for power 75-90%', () => {
      expect(getZone(0.75).name).toBe('Z3');
      expect(getZone(0.85).name).toBe('Z3');
      expect(getZone(0.89).name).toBe('Z3');
    });

    it('returns Z4 for power 90-105%', () => {
      expect(getZone(0.90).name).toBe('Z4');
      expect(getZone(1.00).name).toBe('Z4');
      expect(getZone(1.04).name).toBe('Z4');
    });

    it('returns Z5 for power 105-120%', () => {
      expect(getZone(1.05).name).toBe('Z5');
      expect(getZone(1.10).name).toBe('Z5');
      expect(getZone(1.19).name).toBe('Z5');
    });

    it('returns Z6 for power > 120%', () => {
      expect(getZone(1.20).name).toBe('Z6');
      expect(getZone(1.50).name).toBe('Z6');
      expect(getZone(2.00).name).toBe('Z6');
    });
  });

  describe('getZoneColor', () => {
    it('returns a color string for any power value', () => {
      expect(getZoneColor(0.40)).toMatch(/^#/);
      expect(getZoneColor(0.65)).toMatch(/^#/);
      expect(getZoneColor(1.10)).toMatch(/^#/);
    });

    it('returns Z1 color for recovery power', () => {
      expect(getZoneColor(0.40)).toBe('#A0C4E8');
    });

    it('returns Z4 color for threshold power', () => {
      expect(getZoneColor(1.00)).toBe('#EAB308');
    });

    it('returns Z6 color for anaerobic power', () => {
      expect(getZoneColor(1.30)).toBe('#EF4444');
    });
  });

  describe('getRampZoneColor', () => {
    it('uses average of start and end power for zone', () => {
      // avg of 0.50 and 1.00 = 0.75, which is Z3 (green)
      expect(getRampZoneColor(0.50, 1.00)).toBe('#22C55E');
    });

    it('handles equal start and end', () => {
      expect(getRampZoneColor(0.65, 0.65)).toBe(getZoneColor(0.65));
    });
  });

  describe('toWatts', () => {
    it('converts fraction to absolute watts', () => {
      expect(toWatts(1.0, 300)).toBe(300);
      expect(toWatts(0.75, 300)).toBe(225);
      expect(toWatts(0.50, 200)).toBe(100);
    });

    it('rounds to nearest watt', () => {
      expect(toWatts(0.333, 100)).toBe(33);
    });
  });

  describe('toPercent', () => {
    it('converts fraction to percentage string', () => {
      expect(toPercent(0.75)).toBe('75%');
      expect(toPercent(1.0)).toBe('100%');
      expect(toPercent(0.50)).toBe('50%');
    });

    it('rounds to nearest integer', () => {
      expect(toPercent(0.756)).toBe('76%');
    });
  });
});
