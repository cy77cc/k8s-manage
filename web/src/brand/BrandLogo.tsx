import React from 'react';
import { brand, type BrandLogoVariant } from './brand';

interface BrandLogoProps {
  variant?: BrandLogoVariant;
  width?: number;
  height?: number;
  alt?: string;
  className?: string;
}

export const BRAND_MIN_SIZE = 20;

const BrandLogo: React.FC<BrandLogoProps> = ({
  variant = 'primary',
  width,
  height,
  alt,
  className,
}) => {
  const defaultWidth = variant === 'primary' ? 132 : 32;
  const safeWidth = Math.max(BRAND_MIN_SIZE, width ?? defaultWidth);
  const safeHeight = Math.max(BRAND_MIN_SIZE, height ?? 32);

  return (
    <img
      src={brand.logos[variant]}
      alt={alt || `${brand.canonicalName} logo`}
      width={safeWidth}
      height={safeHeight}
      className={className}
      style={{
        minWidth: BRAND_MIN_SIZE,
        minHeight: BRAND_MIN_SIZE,
        objectFit: 'contain',
      }}
    />
  );
};

export default BrandLogo;
