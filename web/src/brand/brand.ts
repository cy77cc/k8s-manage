import logoMonochrome from '../assets/brand/logo-monochrome.svg';
import logoPrimary from '../assets/brand/logo-primary.svg';
import logoSimplified from '../assets/brand/logo-simplified.svg';

export type BrandLogoVariant = 'primary' | 'simplified' | 'monochrome';

export const brandCandidates = ['NebulaOps', 'CloudHelm', 'SkyForge'];

export const brand = {
  canonicalName: 'NebulaOps',
  tagline: '云原生运维控制台',
  logos: {
    primary: logoPrimary,
    simplified: logoSimplified,
    monochrome: logoMonochrome,
  } as Record<BrandLogoVariant, string>,
};

export const getAppTitle = (suffix?: string) => {
  if (!suffix) return brand.canonicalName;
  return `${suffix} · ${brand.canonicalName}`;
};
